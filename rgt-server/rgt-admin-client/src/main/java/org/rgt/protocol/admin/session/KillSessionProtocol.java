/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;
import static org.rgt.admin.AdminOperation.KILL_SESSION;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = KILL_SESSION)
public class KillSessionProtocol extends AbstractProtocol<AdminOperation, KillSessionRequest, Response>{

   public KillSessionProtocol() {
      super(KILL_SESSION, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(KillSessionRequest request, ByteArrayBuffer buffer) {
      buffer.putLong(request.sessionId());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
