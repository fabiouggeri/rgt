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
import org.rgt.protocol.Request;
import static org.rgt.admin.AdminOperation.GET_CONFIG;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = GET_CONFIG)
public class GetConfigProtocol extends AbstractProtocol<AdminOperation, Request, GetConfigResponse> {

   public GetConfigProtocol() {
      super(GET_CONFIG, (c) -> new GetConfigResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetConfigResponse response) {
      final int count = buffer.getInt();
      final HashMap<String, String> values = new HashMap<>(count);
      for (int i = 0; i < count; i++) {
         values.put(buffer.getString(), buffer.getString());
      }
      response.config(values);
      values.clear();
   }
}
