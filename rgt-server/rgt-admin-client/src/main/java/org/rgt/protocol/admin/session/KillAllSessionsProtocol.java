/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Request;
import static org.rgt.admin.AdminOperation.KILL_ALL_SESSIONS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = KILL_ALL_SESSIONS)
public class KillAllSessionsProtocol extends AbstractProtocol<AdminOperation, Request, KillAllSessionsResponse> {

   public KillAllSessionsProtocol() {
      super(KILL_ALL_SESSIONS, (c) -> new KillAllSessionsResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, KillAllSessionsResponse response) {
      response.killedSessions(buffer.getInt());
   }
}
