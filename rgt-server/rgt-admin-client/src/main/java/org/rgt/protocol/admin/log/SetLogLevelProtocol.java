/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.log;

import org.rgt.ByteArrayBuffer;
import org.rgt.RGTLogLevel;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;
import static org.rgt.admin.AdminOperation.SET_LOG_LEVEL;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = SET_LOG_LEVEL)
public class SetLogLevelProtocol extends AbstractProtocol<AdminOperation, SetLogLevelRequest, Response> {

   public SetLogLevelProtocol() {
      super(SET_LOG_LEVEL, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(SetLogLevelRequest request, ByteArrayBuffer buffer) {
      buffer.putString(request.appLogLevel().name())
              .putString(request.serverLogLevel().name())
              .putString(request.teLogLevel().name());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
