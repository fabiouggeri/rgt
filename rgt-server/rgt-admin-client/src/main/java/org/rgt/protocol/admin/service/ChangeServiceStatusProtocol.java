/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.service;

import org.rgt.ByteArrayBuffer;
import org.rgt.ServerStatus;
import org.rgt.protocol.AdminProtocolRegister;
import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.START_SERVICE;
import static org.rgt.admin.AdminOperation.STOP_SERVICE;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = {START_SERVICE, STOP_SERVICE})
public class ChangeServiceStatusProtocol extends AbstractProtocol<AdminOperation, Request, ChangeServiceStatusResponse> {

   public ChangeServiceStatusProtocol(AdminOperation operation) {
      super(operation, (c) -> new ChangeServiceStatusResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(Request request, ByteArrayBuffer buffer) {
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, ChangeServiceStatusResponse response) {
      response.status(ServerStatus.valueOf(buffer.getString()));
   }
}
