/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.server;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.GET_STATS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio
 */
@AdminProtocolRegister(operation = GET_STATS, fromVersion = 7)
public class ServerStatsProtocol extends AbstractProtocol<AdminOperation, Request, ServerStatsResponse> {

   public ServerStatsProtocol() {
      super(GET_STATS, (c) -> new ServerStatsResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, ServerStatsResponse response) {
      response.setBytesReceived(buffer.getLong());
      response.setBytesSent(buffer.getLong());
      response.setPacketsReceived(buffer.getLong());
      response.setPacketsSent(buffer.getLong());
   }
}
