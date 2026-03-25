/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.ByteArrayBuffer;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.GET_SESSION_STATS;
import org.rgt.admin.AdminResponseCode;
import org.rgt.client.SessionStats;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;

/**
 *
 * @author fabio
 */
@AdminProtocolRegister(operation = GET_SESSION_STATS, fromVersion = 7)
public class GetSessionStatsProtocol extends AbstractProtocol<AdminOperation, GetSessionStatsRequest, GetSessionStatsResponse> {
   
   public GetSessionStatsProtocol() {
      super(GET_SESSION_STATS, (c) -> new GetSessionStatsResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(GetSessionStatsRequest request, ByteArrayBuffer buffer) {
      buffer.putLong(request.sessionId());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetSessionStatsResponse response) {
      final SessionStats stats = new SessionStats();
     // te stats
      stats.setBytesReceivedFromTerminal(buffer.getLong());
      stats.setBytesSentToTerminal(buffer.getLong());
      stats.setPacketsReceivedFromTerminal(buffer.getLong());
      stats.setPacketsSentToTerminal(buffer.getLong());
      // app stats
      stats.setBytesReceivedFromApplication(buffer.getLong());
      stats.setBytesSentToApplication(buffer.getLong());
      stats.setPacketsReceivedFromApplication(buffer.getLong());
      stats.setPacketsSentToApplication(buffer.getLong());
      response.setSessionStats(stats);
   }
   
}
