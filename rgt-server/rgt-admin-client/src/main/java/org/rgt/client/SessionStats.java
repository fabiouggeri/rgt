/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.client;

/**
 *
 * @author fabio
 */
public class SessionStats {

   private long bytesReceivedFromTerminal;
   private long bytesSentToTerminal;
   private long packetsReceivedFromTerminal;
   private long packetsSentToTerminal;
   private long bytesReceivedFromApplication;
   private long bytesSentToApplication;
   private long packetsReceivedFromApplication;
   private long packetsSentToApplication;

   public SessionStats(long bytesReceived, long bytesSent, long packetsReceived, long packetsSent) {
      this.bytesReceivedFromTerminal = bytesReceived;
      this.bytesSentToTerminal = bytesSent;
      this.packetsReceivedFromTerminal = packetsReceived;
      this.packetsSentToTerminal = packetsSent;
   }

   public SessionStats(final SessionStats other) {
      this.bytesReceivedFromTerminal = other.bytesReceivedFromTerminal;
      this.bytesSentToTerminal = other.bytesSentToTerminal;
      this.packetsReceivedFromTerminal = other.packetsReceivedFromTerminal;
      this.packetsSentToTerminal = other.packetsSentToTerminal;
   }

   public SessionStats() {
   }

   public long getBytesReceivedFromTerminal() {
      return bytesReceivedFromTerminal;
   }

   public void setBytesReceivedFromTerminal(long bytesReceivedFromTerminal) {
      this.bytesReceivedFromTerminal = bytesReceivedFromTerminal;
   }

   public long getBytesSentToTerminal() {
      return bytesSentToTerminal;
   }

   public void setBytesSentToTerminal(long bytesSentToTerminal) {
      this.bytesSentToTerminal = bytesSentToTerminal;
   }

   public long getPacketsReceivedFromTerminal() {
      return packetsReceivedFromTerminal;
   }

   public void setPacketsReceivedFromTerminal(long packetsReceivedFromTerminal) {
      this.packetsReceivedFromTerminal = packetsReceivedFromTerminal;
   }

   public long getPacketsSentToTerminal() {
      return packetsSentToTerminal;
   }

   public void setPacketsSentToTerminal(long packetsSentToTerminal) {
      this.packetsSentToTerminal = packetsSentToTerminal;
   }

   /**
    * @return the bytesReceivedFromApplication
    */
   public long getBytesReceivedFromApplication() {
      return bytesReceivedFromApplication;
   }

   /**
    * @param bytesReceivedFromApplication the bytesReceivedFromApplication to set
    */
   public void setBytesReceivedFromApplication(long bytesReceivedFromApplication) {
      this.bytesReceivedFromApplication = bytesReceivedFromApplication;
   }

   /**
    * @return the bytesSentToApplication
    */
   public long getBytesSentToApplication() {
      return bytesSentToApplication;
   }

   /**
    * @param bytesSentToApplication the bytesSentToApplication to set
    */
   public void setBytesSentToApplication(long bytesSentToApplication) {
      this.bytesSentToApplication = bytesSentToApplication;
   }

   /**
    * @return the packetsReceivedFromApplication
    */
   public long getPacketsReceivedFromApplication() {
      return packetsReceivedFromApplication;
   }

   /**
    * @param packetsReceivedFromApplication the packetsReceivedFromApplication to set
    */
   public void setPacketsReceivedFromApplication(long packetsReceivedFromApplication) {
      this.packetsReceivedFromApplication = packetsReceivedFromApplication;
   }

   /**
    * @return the packetsSentToApplication
    */
   public long getPacketsSentToApplication() {
      return packetsSentToApplication;
   }

   /**
    * @param packetsSentToApplication the packetsSentToApplication to set
    */
   public void setPacketsSentToApplication(long packetsSentToApplication) {
      this.packetsSentToApplication = packetsSentToApplication;
   }

}
