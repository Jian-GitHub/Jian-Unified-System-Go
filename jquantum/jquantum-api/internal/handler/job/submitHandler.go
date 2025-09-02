package job

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"jian-unified-system/jquantum/jquantum-api/internal/logic/job"
	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"
	"net/http"
)

const TARGET_DIR = "/harmoniacore/jquantum/data/user" // 指定要保存的路径

func SubmitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := job.NewSubmitLogic(r.Context(), svcCtx)

		//jobID := uuid.NewString()
		data := checkFiles(w, r)

		resp, err := l.Submit(data)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func checkFiles(w http.ResponseWriter, r *http.Request) []byte {
	// 解析 multipart form
	err := r.ParseMultipartForm(32 << 20) // 最大 32MB
	if err != nil {
		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
			Code:    -1,
			Message: "No file uploaded: " + err.Error(),
		})
	}
	file, handler, err := r.FormFile("ariadne")
	if handler.Filename != "thread.zip" {
		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
			Code:    -2,
			Message: "file does not follow the rule",
		})
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
			Code:    -3,
			Message: err.Error(),
		})
		return nil
	}

	return data

	//id, err := r.Context().Value("id").(json.Number).Int64()
	//if err != nil {
	//	httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//		Code:    -3,
	//		Message: "No id: " + err.Error(),
	//	})
	//}

	////dir := filepath.Join(TARGET_DIR, strconv.FormatInt(id, 10), jobID)
	//
	//// 创建目标目录
	//if err := os.MkdirAll(dir, os.ModePerm); err != nil {
	//	httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//		Code:    -4,
	//		Message: "Failed to create target dir: " + err.Error(),
	//	})
	//	return
	//}
	//
	//// 读取 zip 文件
	//zipReader, err := zip.NewReader(file, handler.Size)
	//if err != nil {
	//	httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//		Code:    -5,
	//		Message: "Invalid zip file: " + err.Error(),
	//	})
	//	return
	//}
	//
	//// 解压每个文件到指定路径
	//for _, zipFile := range zipReader.File {
	//	zipFilePath := filepath.Join(dir, zipFile.Name)
	//
	//	if zipFile.FileInfo().IsDir() {
	//		_ = os.MkdirAll(zipFilePath, os.ModePerm)
	//		continue
	//	}
	//
	//	// 确保目录存在
	//	if err := os.MkdirAll(filepath.Dir(zipFilePath), os.ModePerm); err != nil {
	//		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//			Code:    -6,
	//			Message: "Failed to create dir for file: " + err.Error(),
	//		})
	//		return
	//	}
	//
	//	// 创建目标文件
	//	dstFile, err := os.OpenFile(zipFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
	//	if err != nil {
	//		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//			Code:    -7,
	//			Message: "Failed to create file (" + zipFile.Name + "): " + err.Error()})
	//		return
	//	}
	//
	//	// 解压内容
	//	srcFile, err := zipFile.Open()
	//	if err != nil {
	//		_ = dstFile.Close()
	//		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//			Code:    -8,
	//			Message: "Failed to open zip entry: (" + zipFile.Name + "): " + err.Error()})
	//		return
	//	}
	//
	//	_, err = io.Copy(dstFile, srcFile)
	//	_ = srcFile.Close()
	//	_ = dstFile.Close()
	//
	//	if err != nil {
	//		httpx.WriteJson(w, http.StatusInternalServerError, types.BaseResponse{
	//			Code:    -9,
	//			Message: "Failed to write file (" + zipFile.Name + "): " + err.Error()})
	//		return
	//	}
	//}
}
