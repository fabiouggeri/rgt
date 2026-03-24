/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.ByteArrayBuffer;
import org.rgt.TerminalUtil;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.GET_FILE;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = GET_FILE)
public class GetFileProtocol extends AbstractProtocol<AdminOperation, GetFileRequest, GetFileResponse> {

   public GetFileProtocol() {
      super(GET_FILE, (c) -> new GetFileResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(GetFileRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.pathFilename());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetFileResponse response) {
      final String fileName = buffer.getString();
      final FileInfo fileInfo = new FileInfo(fileName);
      response.fileInfo(fileInfo);
      if (!TerminalUtil.isEmpty(fileName)) {
         fileInfo.setLength(buffer.getLong());
         fileInfo.setCreationTime(buffer.getDate().getTime());
         fileInfo.setLastModificationTime(buffer.getDate().getTime());
      }
      response.data(buffer.getArray());
   }

}
