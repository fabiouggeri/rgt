/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.client;

import java.io.IOException;
import java.net.Socket;
import java.nio.ByteBuffer;

/**
 *
 * @author fabio_uggeri
 */
public class SocketConnection implements Connection {
   
   private final Socket socket;

   public static SocketConnection open(final String host, final int port) throws IOException {
      return new SocketConnection(host, port);
   }

   private SocketConnection(final String host, final int port) throws IOException {
      this.socket = new Socket(host, port);
      this.socket.setTcpNoDelay(true);
      this.socket.setKeepAlive(true);
      this.socket.setSoTimeout(DEFAULT_IO_TIMEOUT);
      this.socket.setSoLinger(true, DEFAULT_CLOSE_TIMEOUT);
   }

   @Override
   public int write(ByteBuffer src) throws IOException {
      final byte buffer[] = new byte[src.remaining()];
      src.get(buffer);
      socket.getOutputStream().write(buffer);
      return buffer.length;
   }

   @Override
   public int read(ByteBuffer dst) throws IOException {
      final byte buffer[] = new byte[dst.remaining()];
      final int read = socket.getInputStream().read(buffer);
      if (read > 0) {
         dst.put(buffer, 0, read);
      }
      return read;
   }

   @Override
   public final void close() throws IOException {
      socket.close();
   }
   
   @Override
   public boolean isConnected() {
      return socket.isConnected();
   }

}
