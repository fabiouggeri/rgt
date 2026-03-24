/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;
import static org.rgt.admin.AdminOperation.REMOVE_USER;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = REMOVE_USER)
public class RemoveUserProtocol extends AbstractProtocol<AdminOperation, RemoveUserRequest, Response> {

   public RemoveUserProtocol() {
      super(REMOVE_USER, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(RemoveUserRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.username());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
