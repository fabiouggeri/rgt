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
import static org.rgt.admin.AdminOperation.ADD_USER;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = ADD_USER)
public class AddUserProtocol extends AbstractProtocol<AdminOperation, AddUserRequest, Response> {

   public AddUserProtocol() {
      super(ADD_USER, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(AddUserRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.user().getUsername())
              .putString(request.user().getPassword())
              .putBoolean(request.user().getExpiration() != null);
      if (request.user().getExpiration() != null) {
         buffer.putDate(request.user().getExpiration());
      }
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
      // Do nothing
   }
}
