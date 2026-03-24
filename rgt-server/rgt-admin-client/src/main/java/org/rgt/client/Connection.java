/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package org.rgt.client;

import java.io.Closeable;
import java.io.IOException;
import java.nio.ByteBuffer;

/**
 *
 * @author fabio_uggeri
 */
public interface Connection extends Closeable {

   public static final int DEFAULT_CLOSE_TIMEOUT = 10000;

   public static final int DEFAULT_IO_TIMEOUT = 60000;

   boolean isConnected();

   int read(ByteBuffer dst) throws IOException;

   int write(ByteBuffer src) throws IOException;
   
}
