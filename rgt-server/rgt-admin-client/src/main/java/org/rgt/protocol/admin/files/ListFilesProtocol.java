/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template fileInfo, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.LIST_FILES;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.admin.files.FileInfo.FileType;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = LIST_FILES)
public class ListFilesProtocol extends AbstractProtocol<AdminOperation, ListFilesRequest, ListFilesResponse> {

   public ListFilesProtocol() {
      super(LIST_FILES, (c) -> new ListFilesResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(ListFilesRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.path());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, ListFilesResponse response) {
      final int filesCount;
      response.folderPathname(buffer.getString());
      filesCount = buffer.getInt();
      for (int i = 0; i < filesCount; i++) {
         response.fileInfo(new FileInfo(buffer.getString(),
                 FileType.byValue(buffer.getByte()),
                 buffer.getLong(),
                 buffer.getDate().getTime(),
                 buffer.getDate().getTime()));
      }
   }

}
