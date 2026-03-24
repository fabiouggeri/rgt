/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.ByteArrayBuffer;
import org.rgt.TextScreen;
import org.rgt.admin.AdminOperation;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.AbstractProtocol;
import org.rgt.protocol.AdminProtocolRegister;
import static org.rgt.admin.AdminOperation.SEND_TERMINAL_REQUEST;
import org.rgt.terminal.TerminalOperation;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = SEND_TERMINAL_REQUEST)
public class GetTerminalScreenProtocol extends AbstractProtocol<AdminOperation, GetTerminalScreenRequest, GetTerminalScreenResponse> {

   public static final byte SCREEN_SIMPLE = 0x00;

   public static final byte SCREEN_EXTENDED = 0x01;

   public GetTerminalScreenProtocol() {
      super(SEND_TERMINAL_REQUEST, (c) -> new GetTerminalScreenResponse(AdminResponseCode.getByValue(c)));
   }

   @Override
   protected void serializeRequest(GetTerminalScreenRequest request, ByteArrayBuffer buffer) {
      buffer.putLong(request.sessionId());
      buffer.put(TerminalOperation.ADMIN_GET_SCREEN.getCode());
   }

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, GetTerminalScreenResponse response) {
      final byte screenType;
      final int rows;
      final int columns;
      final TextScreen screen;
      //Charset charset = Charset.forName("IBM-850");

      screenType = buffer.getByte();
      rows = buffer.getShort();
      columns = buffer.getShort();
      screen = new TextScreen(rows, columns);
      screen.setCursorStyle(buffer.getShort());
      screen.setCursorRow(buffer.getShort());
      screen.setCursorCol(buffer.getShort());
      if (screenType == SCREEN_SIMPLE) {
         for (int row = 0; row < rows; row++) {
            for (int col = 0; col < columns; col++) {
               screen.put(row, col, (char)buffer.getByte(), buffer.getByte());
            }
         }
      } else {
         for (int row = 0; row < rows; row++) {
            for (int col = 0; col < columns; col++) {
               screen.put(row, col, (char) (buffer.getShort() & 0xFFFF), buffer.getByte());
               buffer.getByte(); // ignore attribute
            }
         }
      }
      response.screen(screen);
   }

}
