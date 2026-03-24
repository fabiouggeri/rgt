/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import javax.swing.event.ChangeEvent;
import javax.swing.table.TableCellEditor;

/**
 *
 * @author fabio_uggeri
 */
public class CustomChangeEvent extends ChangeEvent {

   private final int row;
   private final int col;

   public CustomChangeEvent(TableCellEditor source, int row, int col) {
      super(source);
      this.row = row;
      this.col = col;
   }

   public int getRow() {
      return row;
   }

   public int getCol() {
      return col;
   }

   public TableCellEditor getCellEditor() {
      return (TableCellEditor)super.getSource();
   }


}
