/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.login;

import org.rgt.ByteArrayBuffer;
import org.rgt.ServerStatus;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.LOGIN;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = LOGIN, fromVersion = 4)
public class LoginProtocolV4 extends AbstractProtocol<AdminOperation, LoginRequest, LoginResponse> {

   public LoginProtocolV4() {
      super(LOGIN, (c) -> new LoginResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(LoginRequest request, ByteArrayBuffer buffer) {
      buffer.putShort(request.protocolVersion())
              .putString(request.username())
              .putString(request.password());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, LoginResponse response) {
      response.serverStatus(ServerStatus.getByName(buffer.getString()))
              .sessionsCount(buffer.getInt())
              .startTime(buffer.getLong())
              .readOnly(buffer.getBoolean());
      // Support versions older than 4
      if (buffer.remaining() >= Short.BYTES) {
         response.adminProtocolVersion(buffer.getShort());
         response.serverVersion(buffer.getString());
         response.userEditing(buffer.getString());
      } else {
         response.adminProtocolVersion((short) 0);
      }
   }
}
