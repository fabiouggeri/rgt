/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import org.rgt.auth.TerminalUser;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Request;
import static org.rgt.admin.AdminOperation.GET_USERS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = GET_USERS)
public class GetUsersProtocol extends AbstractProtocol<AdminOperation, Request, GetUsersResponse> {

   public GetUsersProtocol() {
      super(GET_USERS, (c) -> new GetUsersResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetUsersResponse response) {
      if (response.isSuccess()) {
         final int usersCount = buffer.getInt();
         response.setUserAuthentication(usersCount >= 0);
         for (int i = 0; i < usersCount; i++) {
            final TerminalUser user = new TerminalUser(buffer.getString(), buffer.getString());
            final boolean expiration = buffer.getBoolean();
            if (expiration) {
               user.setExpiration(buffer.getDate().getTime());
            }
         }
      }
   }
}
