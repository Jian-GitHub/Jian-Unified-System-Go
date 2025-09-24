package deploy

import (
	"context"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"math"
	"os"
	"path/filepath"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClusterResource 集群资源
type ClusterResource struct {
	TotalCPU       int64
	TotalMem       int64
	TotalSlots     int64
	TotalSlotsPow2 int64
	MaxQubits      int64
	Nodes          map[string]NodeResource
}

// NodeResource 节点资源
type NodeResource struct {
	CPU   int64
	Mem   int64
	IP    string
	Host  string
	Slots int64
}

type K8sDeployService struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewK8sDeployService 创建 service
func NewK8sDeployService(namespace string) (K8sDeployService, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return K8sDeployService{}, fmt.Errorf("create clientset failed: %v", err)
	}
	return K8sDeployService{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// GetK8sClient 尝试自动选择本地 or 集群内的配置
func GetK8sClient() (*kubernetes.Clientset, error) {
	// 先尝试 InCluster
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster k8s client: %w", err)
		}
		return clientset, nil
	}

	// 如果 InClusterConfig 失败，就 fallback 到本地 kubeconfig
	home := os.Getenv("HOME")
	var kubeconfig string
	if home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, fmt.Errorf("no kubeconfig file found at %s", kubeconfig)
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build local kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create local k8s client: %w", err)
	}
	return clientset, nil
}

// CollectClusterResource 统计集群资源
func (s *K8sDeployService) CollectClusterResource() (*ClusterResource, error) {
	ctx := context.Background()

	// 获取所有 app=jquantum-rpc 的 Pod
	pods, err := s.clientset.CoreV1().Pods(s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=jquantum-rpc",
	})
	if err != nil {
		return nil, fmt.Errorf("list pods failed: %v", err)
	}

	// 获取 Node 列表，用于统计 CPU / Mem
	nodes, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes failed: %v", err)
	}

	result := &ClusterResource{
		Nodes: make(map[string]NodeResource),
	}

	const reserveMem int64 = 2 << 30 // 2GiB

	// 先统计节点资源
	nodeCPUMap := make(map[string]int64)
	nodeMemMap := make(map[string]int64)
	for _, node := range nodes.Items {
		cpuQty := node.Status.Capacity[corev1.ResourceCPU]
		memQty := node.Status.Capacity[corev1.ResourceMemory]

		cpuCount, _ := strconv.ParseInt(cpuQty.AsDec().String(), 10, 64)
		memAlloc := memQty.Value()
		memUsed := memAlloc - reserveMem
		if memUsed <= 0 {
			memUsed = int64(float64(memAlloc) * 0.8)
		}

		nodeCPUMap[node.Name] = cpuCount
		nodeMemMap[node.Name] = memUsed

		result.TotalCPU += cpuCount
		result.TotalMem += memUsed
	}

	// 遍历 Pod，获取 IP 并计算 slots
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning || pod.Status.PodIP == "" {
			continue
		}
		nodeName := pod.Spec.NodeName
		cpu, ok := nodeCPUMap[nodeName]
		if !ok {
			continue
		}

		slots := cpu - 1
		if slots < 1 {
			slots = 1
		}

		result.Nodes[pod.Name] = NodeResource{
			CPU:   cpu,
			Mem:   nodeMemMap[nodeName],
			IP:    pod.Status.PodIP,
			Host:  nodeName,
			Slots: slots,
		}

		result.TotalSlots += slots
	}

	// 总 slots 取 2 的幂次
	if result.TotalSlots > 0 {
		result.TotalSlotsPow2 = int64(math.Pow(2, math.Floor(math.Log2(float64(result.TotalSlots)))))
	}

	// 最大可支持 qubits 数
	if result.TotalMem > 0 {
		const bytesPerComplex = int64(16)
		maxStates := float64(result.TotalMem / bytesPerComplex)
		result.MaxQubits = int64(math.Floor(math.Log2(maxStates)))
	}

	return result, nil
}

// GenerateHostsFile 生成 MPI hosts 文件，使用每个 Pod 的 IP 和 slots
func (s *K8sDeployService) GenerateHostsFile(resource *ClusterResource, filePath string) error {
	var content string
	for _, podRes := range resource.Nodes { // 这里的 key 是 Pod 名称
		if podRes.IP == "" || podRes.Slots == 0 {
			continue
		}
		content += fmt.Sprintf("%s slots=%d\n", podRes.IP, podRes.Slots)
	}

	if content == "" {
		return fmt.Errorf("no valid pods for hosts file")
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write hosts file failed: %v", err)
	}

	return nil
}
