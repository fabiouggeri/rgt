/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.server;

import org.rgt.ByteArrayBuffer;
import org.rgt.ServerStatus;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.AdminProtocolRegister;
import static org.rgt.admin.AdminOperation.GET_STATUS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = GET_STATUS)
public class ServerInfoProtocol extends AbstractProtocol<AdminOperation, Request, ServerInfoResponse> {

   public ServerInfoProtocol() {
      super(GET_STATUS, (c) -> new ServerInfoResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, ServerInfoResponse response) {
      response.status(ServerStatus.getByName(buffer.getString()))
              .sessionsCount(buffer.getInt())
              .startTime(buffer.getLong());
   }

}
