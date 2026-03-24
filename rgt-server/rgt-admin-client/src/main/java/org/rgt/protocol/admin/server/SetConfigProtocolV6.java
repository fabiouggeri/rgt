/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.server;

import org.rgt.ByteArrayBuffer;
import static org.rgt.admin.AdminOperation.SET_CONFIG;
import org.rgt.protocol.AdminProtocolRegister;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = SET_CONFIG, fromVersion = 6)
public class SetConfigProtocolV6 extends SetConfigProtocol {

   @Override
   protected void serializeRequest(SetConfigRequest request, ByteArrayBuffer buffer) {
      buffer.putBoolean(request.removeMissingOptions());
      buffer.putInt(request.config().size());
      request.config().forEach((k, v) -> buffer.putString(k).putString(v));
   }
}
