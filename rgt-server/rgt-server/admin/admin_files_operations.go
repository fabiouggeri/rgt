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
	adminProtocol.Operation(ADM_LIST_FILES, "List files").Version(4).Executor(listFiles)
	adminProtocol.Operation(ADM_GET_FILE, "Get file").Version(4).Executor(getFile)
	adminProtocol.Operation(ADM_PUT_FILE, "Put file").Version(4).Executor(putFile)
	adminProtocol.Operation(ADM_REMOVE_FILE, "Remove file").Version(4).Executor(removeFile)
}

func (r *ListFilesRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.path = buf.GetString()
}

func (r *ListFilesRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.path)
}

func (r *ListFilesResponse) FromBuffer(buf *buffer.ByteBuffer) {
	folderName := buf.GetString()
	filesCount := int(buf.GetInt32())
	r.folderPathname = folderName
	r.filesInfo = make([]FileInfo, filesCount)
	for _, f := range r.filesInfo {
		f.name = buf.GetString()
		f.fileType = FileType(buf.GetInt8())
		f.length = buf.GetInt64()
		f.creationTime = buf.GetDate()
		f.lastModificationTime = buf.GetDate()
	}
}

func (r *ListFilesResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.folderPathname)
	buf.PutInt32(int32(len(r.filesInfo)))
	for _, f := range r.filesInfo {
		buf.PutString(f.name)
		buf.PutInt8(int8(f.fileType))
		buf.PutInt64(f.length)
		buf.PutDate(f.creationTime)
		buf.PutDate(f.lastModificationTime)
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

func listFiles(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.listFiles()")
	req := &ListFilesRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
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
	resp := &ListFilesResponse{
		folderPathname: path,
		filesInfo:      make([]FileInfo, len(entries)),
	}
	for i, dirEntry := range entries {
		dirEntryInfo, infoErr := dirEntry.Info()
		if infoErr == nil {
			fileToFileInfo(dirEntryInfo, &resp.filesInfo[i])
		}
	}
	bufResp := buffer.NewCapacity(uint32((len(entries) * 128) + len(path) + 4))
	protocol.PutResponse(resp, bufResp)
	return bufResp, nil
}

func (r *GetFileRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.filePathName = buf.GetString()
}

func (r *GetFileRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.filePathName)
}

func (r *GetFileResponse) FromBuffer(buf *buffer.ByteBuffer) {
	fileName := buf.GetString()
	r.fileInfo.name = fileName
	if len(fileName) > 0 {
		r.fileInfo.length = buf.GetInt64()
		r.fileInfo.creationTime = buf.GetDate()
		r.fileInfo.lastModificationTime = buf.GetDate()
	}
	r.data = buf.GetBytes()
}

func (r *GetFileResponse) ToBuffer(buf *buffer.ByteBuffer) {
	if len(r.fileInfo.name) > 0 {
		buf.PutString(r.fileInfo.name)
		buf.PutInt64(r.fileInfo.length)
		buf.PutDate(r.fileInfo.creationTime)
		buf.PutDate(r.fileInfo.lastModificationTime)
	} else {
		buf.PutString("")
	}
	buf.PutSlice(r.data)
}

func getFile(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.getFile()")
	req := &GetFileRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
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
		protocol.PutResponse(resp, bufResp)
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
			return protocol.SuccessResponse(), nil
		} else if nextPack.operation != ADM_GET_FILE {
			return nil, NewError(PROTOCOL_ERROR, "invalid operation received in get file")
		}
	}
	return nil, nil
}

func (r *PutFileRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.filePathName = buf.GetString()
	if len(r.filePathName) > 0 {
		r.fileSize = buf.GetInt64()
		r.force = buf.GetBool()
		r.creationTime = buf.GetDate()
		r.lastModificationTime = buf.GetDate()
	}
	r.data = buf.GetSlice()
}

func (r *PutFileRequest) ToBuffer(buf *buffer.ByteBuffer) {
	if len(r.filePathName) > 0 {
		buf.PutString(r.filePathName)
		buf.PutInt64(r.fileSize)
		buf.PutBool(r.force)
		buf.PutDate(r.creationTime)
		buf.PutDate(r.lastModificationTime)
	} else {
		buf.PutString("")
	}
	buf.PutSlice(r.data)
}

func putFile(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.putFile()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	req := &PutFileRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
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
		err := pack.handler.sendResponse(protocol.SuccessResponse())
		if err != nil {
			log.Debug("[ADMIN] error sending response in PUT_FILE operation: ", err)
			return nil, nil
		}
		nextPack, packErr := pack.handler.readPacket()
		if packErr != nil {
			return nil, NewError(SERVER_ERROR, "error reading new packet: ", packErr)
		}
		if nextPack.operation == ADM_CANCEL {
			return protocol.SuccessResponse(), nil
		} else if nextPack.operation != ADM_PUT_FILE {
			return nil, NewError(PROTOCOL_ERROR, "invalid operation received in get file")
		}
		req = &PutFileRequest{}
		req.FromBuffer(buffer.Wrap(nextPack.body))
		fileSize -= int64(len(req.data))
		_, writeErr := writer.Write(req.data)
		if writeErr != nil {
			return nil, NewError(SERVER_ERROR, "Error writing to file ", req.filePathName)
		}
	}
	return protocol.SuccessResponse(), nil
}

func (r *RemoveFileRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.remotePathName = buf.GetString()
	r.fileName = buf.GetString()
}

func (r *RemoveFileRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.remotePathName)
	buf.PutString(r.fileName)
}

func removeFile(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_files_operations.removeFile()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	req := &RemoveFileRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
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
	return protocol.SuccessResponse(), nil
}
