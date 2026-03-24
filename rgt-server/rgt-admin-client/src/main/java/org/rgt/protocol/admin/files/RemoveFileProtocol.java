/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.REMOVE_FILE;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = REMOVE_FILE)
public class RemoveFileProtocol extends AbstractProtocol<AdminOperation, RemoveFileRequest, Response> {

   public RemoveFileProtocol() {
      super(REMOVE_FILE, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }


   @Override
   protected void serializeRequest(RemoveFileRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.remotePathname()).putString(request.filename());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }

}