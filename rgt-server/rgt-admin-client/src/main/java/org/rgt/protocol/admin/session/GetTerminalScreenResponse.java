/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.TextScreen;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
public class GetTerminalScreenResponse extends Response {

   private TextScreen screen;

   public GetTerminalScreenResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetTerminalScreenResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public TextScreen screen() {
      return screen;
   }

   public GetTerminalScreenResponse screen(final TextScreen screen) {
      this.screen = screen;
      return this;
   }
}
