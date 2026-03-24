/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.client;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.StandardSocketOptions;
import java.nio.ByteBuffer;
import java.nio.channels.SocketChannel;

/**
 *
 * @author fabio_uggeri
 */
public class SocketChannelConnection implements Connection {
   
   private final SocketChannel channel;

   public static SocketChannelConnection open(final String host, final int port) throws IOException {
      return new SocketChannelConnection(new InetSocketAddress(host, port));
   }

   private SocketChannelConnection(InetSocketAddress addr) throws IOException {
      this.channel = SocketChannel.open(addr);
      this.channel.configureBlocking(false);
      this.channel.setOption(StandardSocketOptions.TCP_NODELAY, true);
      this.channel.setOption(StandardSocketOptions.SO_KEEPALIVE, true);
   }

   @Override
   public int write(ByteBuffer src) throws IOException {
      return channel.write(src);
   }

   @Override
   public int read(ByteBuffer dst) throws IOException {
      return channel.read(dst);
   }

   @Override
   public final void close() throws IOException {
      channel.close();
   }
   
   @Override
   public boolean isConnected() {
      return channel.isConnected();
   }

}
