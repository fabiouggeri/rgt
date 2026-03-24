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
import static org.rgt.admin.AdminOperation.SET_USERS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = SET_USERS)
public class SetUsersProtocol extends AbstractProtocol<AdminOperation, SetUsersRequest, Response> {

   public SetUsersProtocol() {
      super(SET_USERS, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(SetUsersRequest request, ByteArrayBuffer buffer) {
      buffer.putInt(request.users().size());
      request.users().forEach(u -> {
         buffer.putString(u.getUsername()).putString(u.getPassword()).putBoolean(u.getExpiration() != null);
         if (u.getExpiration() != null) {
            buffer.putDate(u.getExpiration());
         }
      });
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
