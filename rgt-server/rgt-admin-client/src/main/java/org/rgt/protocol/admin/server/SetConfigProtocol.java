/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.server;

import java.util.HashMap;
import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Response;
import static org.rgt.admin.AdminOperation.SET_CONFIG;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = SET_CONFIG)
public class SetConfigProtocol extends AbstractProtocol<AdminOperation, SetConfigRequest, Response>{

   public SetConfigProtocol() {
      super(SET_CONFIG, (c) -> new Response(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(SetConfigRequest request, ByteArrayBuffer buffer) {
      buffer.putInt(request.config().size());
      request.config().forEach((k, v) -> buffer.putString(k).putString(v));
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, Response response) {
   }
}
