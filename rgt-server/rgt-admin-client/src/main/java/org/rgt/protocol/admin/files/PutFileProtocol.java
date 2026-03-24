/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.ByteArrayBuffer;
import org.rgt.TerminalUtil;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.PUT_FILE;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = PUT_FILE)
public class PutFileProtocol extends AbstractProtocol<AdminOperation, PutFileRequest, Response> {

   public PutFileProtocol() {
      super(PUT_FILE, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(PutFileRequest request, ByteArrayBuffer buffer) {
      if (!TerminalUtil.isEmpty(request.filePathname())) {
         buffer.putString(request.filePathname())
                 .putLong(request.fileSize())
                 .putBoolean(request.force())
                 .putDate(request.creationTime())
                 .putDate(request.lastModificationTime())
                 .put(request.data());
      } else {
         buffer.putString("").put(request.data());
      }
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
