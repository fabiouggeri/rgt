/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.service;

import org.rgt.ServerStatus;
import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class ChangeServiceStatusResponse extends Response {

   private ServerStatus status;

   public ChangeServiceStatusResponse() {
      this(AdminResponseCode.SUCCESS, null);
   }

   public ChangeServiceStatusResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public ChangeServiceStatusResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public ChangeServiceStatusResponse(final ServerStatus status) {
      this();
      this.status = status;
   }

   public ServerStatus status() {
      return status;
   }

   public ChangeServiceStatusResponse status(ServerStatus status) {
      this.status = status;
      return this;
   }
}
