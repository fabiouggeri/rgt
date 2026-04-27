package admin

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/util"
	"time"

	"github.com/djherbis/times"
)

type FileType int8

type FileInfo struct {
	name                 string
	length               int64
	creationTime         time.Time
	lastModificationTime time.Time
	fileType             FileType
}

type ListFilesRequest struct {
	protocol.BaseRequest
	path string
}

type ListFilesResponse struct {
	protocol.BaseResponse
	folderPathname string
	filesInfo      []FileInfo
}

type GetFileRequest struct {
	protocol.BaseRequest
	filePathName string
}

type GetFileResponse struct {
	protocol.BaseResponse
	fileInfo FileInfo
	data     []byte
}

type PutFileRequest struct {
	protocol.BaseRequest
	filePathName         string
	creationTime         time.Time
	lastModificationTime time.Time
	fileSize             int64
	data                 []byte
	force                bool
}

type RemoveFileRequest struct {
	protocol.BaseRequest
	remotePathName string
	fileName       string
}

const (
	FILE_TYPE FileType = 0
	DIR_TYPE  FileType = 1
)

func init() {
	registerOperation(ADM_LIST_FILES, listFiles)
	registerOperation(ADM_GET_FILE, getFile)
	registerOperation(ADM_PUT_FILE, putFile)
	registerOperation(ADM_REMOVE_FILE, removeFile)
	registerProtocol(ADM_LIST_FILES, 4, protocol.New(bufferToListFilesRequest, listFilesRequestToBuffer, bufferToListFilesResponse, listFilesResponseToBuffer))
	registerProtocol(ADM_LIST_FILES, 5, protocol.New(bufferToListFilesRequest, listFilesRequestToBuffer, bufferToListFilesResponseV5, listFilesResponseToBufferV5))
	registerProtocol(ADM_GET_FILE, 4, protocol.New(bufferToGetFileRequest, getFileRequestToBuffer, bufferToGetFileResponse, getFileResponseToBuffer))
	registerProtocol(ADM_PUT_FILE, 4, protocol.New(bufferToPutFileRequest, putFileRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_REMOVE_FILE, 4, protocol.New(bufferToRemoveFileRequest, removeFileRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
}

func bufferToListFilesRequest(buf *buffer.ByteBuffer) *ListFilesRequest {
	return &ListFilesRequest{path: buf.GetString()}
}

func listFilesRequestToBuffer(req *ListFilesRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.path)
}

func bufferToListFilesResponse(buf *buffer.ByteBuffer) *ListFilesResponse {
	folderName := buf.GetString()
	filesCount := int(buf.GetInt32())
	resp := &ListFilesResponse{folderPathname: folderName,
		filesInfo: make([]FileInfo, filesCount)}
	for _, f := range resp.filesInfo {
		f.name = buf.GetString()
		f.fileType = FileType(buf.GetInt8())
		f.length = buf.GetInt64()
		f.creationTime = buf.GetDate()
		f.lastModificationTime = buf.GetDate()
	}
	return resp
}

func listFilesResponseToBuffer(resp *ListFilesResponse, buf *buffer.ByteBuffer) {
	buf.PutString(resp.folderPathname)
	buf.PutInt32(int32(len(resp.filesInfo)))
	for _, f := range resp.filesInfo {
		buf.PutString(f.name)
		buf.PutInt8(int8(f.fileType))
		buf.PutInt64(f.length)
		buf.PutDate(f.creationTime)
		buf.PutDate(f.lastModificationTime)
	}
}

func bufferToListFilesResponseV5(buf *buffer.ByteBuffer) *ListFilesResponse {
	folderName := buf.GetString()
	filesCount := int(buf.GetInt32())
	resp := &ListFilesResponse{
		folderPathname: folderName,
		filesInfo:      make([]FileInfo, filesCount),
	}
	for _, f := range resp.filesInfo {
		f.name = buf.GetString()
		f.fileType = FileType(buf.GetInt8())
		f.length = buf.GetInt64()
		f.creationTime = buf.GetDateTime()
		f.lastModificationTime = buf.GetDateTime()
	}
	return resp
}

func listFilesResponseToBufferV5(resp *ListFilesResponse, buf *buffer.ByteBuffer) {
	buf.PutString(resp.folderPathname)
	buf.PutInt32(int32(len(resp.filesInfo)))
	for _, f := range resp.filesInfo {
		buf.PutString(f.name)
		buf.PutInt8(int8(f.fileType))
		buf.PutInt64(f.length)
		buf.PutDateTime(f.creationTime)
		buf.PutDateTime(f.lastModificationTime)
	}
}

func fileToFileInfo(dirEntryInfo os.FileInfo, fileInfo *FileInfo) {
	fileTimes := times.Get(dirEntryInfo)
	fileInfo.name = dirEntryInfo.Name()
	if dirEntryInfo.IsDir() {
		fileInfo.fileType = DIR_TYPE
	} else {
		fileInfo.fileType = FILE_TYPE
	}
	fileInfo.length = dirEntryInfo.Size()
	if fileTimes.HasBirthTime() {
		fileInfo.creationTime = fileTimes.BirthTime()
	} else {
		fileInfo.creationTime = time.UnixMilli(0)
	}
	fileInfo.lastModificationTime = dirEntryInfo.ModTime()
}

func listFiles(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.listFiles()")
	proto, err := findProtocol[*ListFilesRequest, *ListFilesResponse](ADM_LIST_FILES, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	dir, dirErr := os.Stat(req.path)
	if dirErr != nil {
		return nil, NewError(SERVER_ERROR, "Path not found  ", req.path)
	}
	if !dir.IsDir() {
		return nil, NewError(SERVER_ERROR, "Path ", req.path, ". is not a directory")
	}
	path, pathErr := filepath.Abs(req.path)
	if pathErr != nil {
		return nil, NewError(SERVER_ERROR, "Error listing files from ", dir.Name(), ". Error: ", pathErr)
	}
	entries, dirErr := os.ReadDir(req.path)
	if dirErr != nil {
		return nil, NewError(SERVER_ERROR, "Error listing files from ", dir.Name(), ". Error: ", dirErr)
	}
	resp := &ListFilesResponse{folderPathname: path, filesInfo: make([]FileInfo, len(entries))}
	for i, dirEntry := range entries {
		dirEntryInfo, infoErr := dirEntry.Info()
		if infoErr == nil {
			fileToFileInfo(dirEntryInfo, &resp.filesInfo[i])
		}
	}
	bufResp := buffer.New()
	proto.PutResponse(resp, bufResp)
	return bufResp, nil
}

func bufferToGetFileRequest(buf *buffer.ByteBuffer) *GetFileRequest {
	return &GetFileRequest{filePathName: buf.GetString()}
}

func getFileRequestToBuffer(req *GetFileRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.filePathName)
}

func bufferToGetFileResponse(buf *buffer.ByteBuffer) *GetFileResponse {
	fileName := buf.GetString()
	fileInfo := &FileInfo{name: fileName}
	if len(fileName) > 0 {
		fileInfo.length = buf.GetInt64()
		fileInfo.creationTime = buf.GetDate()
		fileInfo.lastModificationTime = buf.GetDate()
	}
	return &GetFileResponse{fileInfo: *fileInfo, data: buf.GetBytes()}
}

func getFileResponseToBuffer(resp *GetFileResponse, buf *buffer.ByteBuffer) {
	if len(resp.fileInfo.name) > 0 {
		buf.PutString(resp.fileInfo.name)
		buf.PutInt64(resp.fileInfo.length)
		buf.PutDate(resp.fileInfo.creationTime)
		buf.PutDate(resp.fileInfo.lastModificationTime)
		buf.PutSlice(resp.data)
	} else {
		buf.PutString("")
		buf.PutSlice(resp.data)
	}

}

func getFile(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.getFile()")
	proto, err := findProtocol[*GetFileRequest, *GetFileResponse](ADM_GET_FILE, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	fileInfo, fileErr := os.Stat(req.filePathName)
	if fileErr != nil {
		return nil, NewError(SERVER_ERROR, "Error getting information for file ", req.filePathName, ". Error: ", fileErr)
	}
	if fileInfo.IsDir() {
		return nil, NewError(SERVER_ERROR, req.filePathName, " is not a file.")
	}
	fileHandle, fileErr := os.Open(req.filePathName)
	if fileErr != nil {
		return nil, NewError(SERVER_ERROR, "Error opening file ", req.filePathName)
	}
	defer fileHandle.Close()
	resp := &GetFileResponse{}
	bufResp := buffer.New()
	fileToFileInfo(fileInfo, &resp.fileInfo)
	fileReader := bufio.NewReader(fileHandle)
	var bufferSize int
	remaining := resp.fileInfo.length
	chunkSize := pack.handler.service.server.Config().AdminFileTransferChunkSize().Get()
	if remaining < int64(chunkSize) {
		bufferSize = int(remaining)
	} else {
		bufferSize = int(chunkSize)
	}
	bufferIO := make([]byte, bufferSize)
	for {
		bytesRead, readErr := fileReader.Read(bufferIO)
		if readErr != nil && readErr != io.EOF {
			return nil, NewError(SERVER_ERROR, "error reading data from file")
		}
		if bytesRead == 0 {
			break
		}
		resp.data = bufferIO[:bytesRead]
		proto.PutResponse(resp, bufResp)
		sendErr := pack.handler.sendResponse(bufResp)
		if sendErr != nil {
			return nil, NewError(SERVER_ERROR, "error sending file chunk: ", sendErr)
		}
		remaining -= int64(bytesRead)
		if remaining <= 0 {
			break
		}
		resp.fileInfo.name = ""
		nextPack, packErr := pack.handler.readPacket()
		if packErr != nil {
			return nil, NewError(SERVER_ERROR, "error reading new packet: ", packErr)
		}
		if nextPack.operation == ADM_CANCEL {
			return SuccessAdminResponse(), nil
		} else if nextPack.operation != ADM_GET_FILE {
			return nil, NewError(PROTOCOL_ERROR, "invalid operation received in get file")
		}
	}
	return nil, nil
}

func bufferToPutFileRequest(buf *buffer.ByteBuffer) *PutFileRequest {
	req := &PutFileRequest{filePathName: buf.GetString()}
	if len(req.filePathName) > 0 {
		req.fileSize = buf.GetInt64()
		req.force = buf.GetBool()
		req.creationTime = buf.GetDate()
		req.lastModificationTime = buf.GetDate()
	}
	req.data = buf.GetSlice()
	return req
}

func putFileRequestToBuffer(req *PutFileRequest, buf *buffer.ByteBuffer) {
	if len(req.filePathName) > 0 {
		buf.PutString(req.filePathName)
		buf.PutInt64(req.fileSize)
		buf.PutBool(req.force)
		buf.PutDate(req.creationTime)
		buf.PutDate(req.lastModificationTime)
	} else {
		buf.PutString("")
	}
	buf.PutSlice(req.data)
}

func putFile(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.putFile()")
	proto, err := findProtocol[*PutFileRequest, *protocol.BaseResponse](ADM_PUT_FILE, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	if len(req.filePathName) == 0 {
		return nil, NewError(SERVER_ERROR, "File path and name not specified")
	}
	if util.FileExists(req.filePathName) {
		if req.force {
			osErr := os.Remove(req.filePathName)
			if osErr != nil {
				return nil, NewError(SERVER_ERROR, "Could not remove file '"+req.filePathName+"'. Error: ", osErr)
			}
		} else {
			return nil, NewError(SERVER_ERROR, "File ", req.filePathName, " already exists")
		}

	}
	fileHandle, osErr := os.Create(req.filePathName)
	if osErr != nil {
		return nil, NewError(SERVER_ERROR, "Error creating file ", req.filePathName)
	}
	defer fileHandle.Close()
	writer := bufio.NewWriter(fileHandle)
	os.Chtimes(req.filePathName, req.creationTime, req.lastModificationTime)
	fileSize := req.fileSize - int64(len(req.data))
	_, writeErr := writer.Write(req.data)
	if writeErr != nil {
		return nil, NewError(SERVER_ERROR, "Error writing to file ", req.filePathName)
	}
	for fileSize > 0 {
		err := pack.handler.sendResponse(SuccessAdminResponse())
		if err != nil {
			log.Debug("[ADMIN] error sending response in PUT_FILE operation: ", err)
			return nil, nil
		}
		nextPack, packErr := pack.handler.readPacket()
		if packErr != nil {
			return nil, NewError(SERVER_ERROR, "error reading new packet: ", packErr)
		}
		if nextPack.operation == ADM_CANCEL {
			return SuccessAdminResponse(), nil
		} else if nextPack.operation != ADM_PUT_FILE {
			return nil, NewError(PROTOCOL_ERROR, "invalid operation received in get file")
		}
		req = proto.GetRequest(buffer.Wrap(nextPack.body))
		fileSize -= int64(len(req.data))
		_, writeErr := writer.Write(req.data)
		if writeErr != nil {
			return nil, NewError(SERVER_ERROR, "Error writing to file ", req.filePathName)
		}
	}
	return SuccessAdminResponse(), nil
}

func bufferToRemoveFileRequest(buf *buffer.ByteBuffer) *RemoveFileRequest {
	return &RemoveFileRequest{remotePathName: buf.GetString(), fileName: buf.GetString()}
}

func removeFileRequestToBuffer(req *RemoveFileRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.remotePathName)
	buf.PutString(req.fileName)
}

func removeFile(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.removeFile()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*RemoveFileRequest, *protocol.BaseResponse](ADM_REMOVE_FILE, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	if len(req.remotePathName) == 0 {
		return nil, NewError(SERVER_ERROR, "Invalid file path '", req.remotePathName, "'")
	} else if len(req.fileName) == 0 {
		return nil, NewError(SERVER_ERROR, "Invalid file name '", req.fileName, "'")
	} else {
		filePathName := path.Join(req.remotePathName, req.fileName)
		info, err := os.Stat(filePathName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, NewError(SERVER_ERROR, "File '", filePathName, "' not found")
			} else {
				return nil, NewError(SERVER_ERROR, "File '", filePathName, "' not found. Error: ", err)
			}
		}
		if info.IsDir() {
			return nil, NewError(SERVER_ERROR, "'"+filePathName+"' is not a file")
		}
		err = os.Remove(filePathName)
		if err != nil {
			return nil, NewError(SERVER_ERROR, "Could not remove file '"+filePathName+"'. Error: ", err)
		}
	}
	return SuccessAdminResponse(), nil
}
