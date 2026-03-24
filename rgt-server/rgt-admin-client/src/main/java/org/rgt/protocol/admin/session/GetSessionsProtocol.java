/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import org.rgt.ByteArrayBuffer;
import org.rgt.SessionStatus;
import org.rgt.admin.AdminOperation;
import org.rgt.client.RemoteSession;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Request;
import static org.rgt.admin.AdminOperation.GET_SESSIONS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = GET_SESSIONS)
public class GetSessionsProtocol extends AbstractProtocol<AdminOperation, Request, GetSessionsResponse> {

   public GetSessionsProtocol() {
      super(GET_SESSIONS, (c) -> new GetSessionsResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetSessionsResponse response) {
      final int sessionsCount = buffer.getInt();
      response = new GetSessionsResponse();
      for (int i = 0; i < sessionsCount; i++) {
         final RemoteSession session = new RemoteSession();
         session.setId(buffer.getLong());
         session.setTerminalAddress(buffer.getString());
         session.setOSUser(buffer.getString());
         session.setAppPid(buffer.getLong());
         session.setStatus(SessionStatus.valueOf(buffer.getString()));
         session.setStartTime(buffer.getLong());
         response.getSessions().add(session);
      }
   }
}
