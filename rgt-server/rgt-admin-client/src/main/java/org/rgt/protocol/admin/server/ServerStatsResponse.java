/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.server;

import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio
 */
public class ServerStatsResponse extends Response {

   private long bytesReceived;
   private long bytesSent;
   private long packetsReceived;
   private long packetsSent;

   public ServerStatsResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public ServerStatsResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public ServerStatsResponse() {
      this(AdminResponseCode.SUCCESS, null);
   }

   /**
    * @return the bytesReceived
    */
   public long getBytesReceived() {
      return bytesReceived;
   }

   /**
    * @param bytesReceived the bytesReceived to set
    */
   public void setBytesReceived(long bytesReceived) {
      this.bytesReceived = bytesReceived;
   }

   /**
    * @return the bytesSent
    */
   public long getBytesSent() {
      return bytesSent;
   }

   /**
    * @param bytesSent the bytesSent to set
    */
   public void setBytesSent(long bytesSent) {
      this.bytesSent = bytesSent;
   }

   /**
    * @return the packetsReceived
    */
   public long getPacketsReceived() {
      return packetsReceived;
   }

   /**
    * @param packetsReceived the packetsReceived to set
    */
   public void setPacketsReceived(long packetsReceived) {
      this.packetsReceived = packetsReceived;
   }

   /**
    * @return the packetsSent
    */
   public long getPacketsSent() {
      return packetsSent;
   }

   /**
    * @param packetsSent the packetsSent to set
    */
   public void setPacketsSent(long packetsSent) {
      this.packetsSent = packetsSent;
   }
}
