/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.log;

import org.rgt.RGTLogLevel;
import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class SetLogLevelRequest extends Request {

   private RGTLogLevel appLogLevel;
   private RGTLogLevel serverLogLevel;
   private RGTLogLevel teLogLevel;

   public SetLogLevelRequest() {
      this(RGTLogLevel.ERROR, RGTLogLevel.ERROR, RGTLogLevel.ERROR);
   }

   public SetLogLevelRequest(RGTLogLevel appLogLevel, RGTLogLevel serverLogLevel, RGTLogLevel teLogLevel) {
      super(AdminOperation.SET_LOG_LEVEL);
      this.appLogLevel = appLogLevel;
      this.serverLogLevel = serverLogLevel;
      this.teLogLevel = teLogLevel;
   }

   public RGTLogLevel appLogLevel() {
      return appLogLevel;
   }

   public SetLogLevelRequest appLogLevel(RGTLogLevel appLogLevel) {
      this.appLogLevel = appLogLevel;
      return this;
   }

   public RGTLogLevel serverLogLevel() {
      return serverLogLevel;
   }

   public SetLogLevelRequest serverLogLevel(RGTLogLevel serverLogLevel) {
      this.serverLogLevel = serverLogLevel;
      return this;
   }

   public RGTLogLevel teLogLevel() {
      return teLogLevel;
   }

   public SetLogLevelRequest teLogLevel(RGTLogLevel teLogLevel) {
      this.teLogLevel = teLogLevel;
      return this;
   }
}
