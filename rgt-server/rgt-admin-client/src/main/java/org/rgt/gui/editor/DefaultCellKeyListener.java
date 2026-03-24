/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import javax.swing.JTable;

/**
 *
 * @author fabio_uggeri
 */
public class DefaultCellKeyListener implements KeyListener {

   private final JTable table;
   private final int keys[];

   public DefaultCellKeyListener(final JTable table, int... keys) {
      this.table = table;
      this.keys = keys;
   }

   public DefaultCellKeyListener(final JTable table) {
      this(table, KeyEvent.VK_ENTER, KeyEvent.VK_ESCAPE, KeyEvent.VK_DOWN, KeyEvent.VK_UP, KeyEvent.VK_TAB);
   }

   @Override
   public void keyTyped(KeyEvent e) {
   }

   @Override
   public void keyPressed(KeyEvent e) {
      if (!e.isConsumed()) {
         for (int key : keys) {
            if (e.getKeyCode() == key) {
               e.consume();
               table.dispatchEvent(new KeyEvent(table, e.getID(), e.getWhen(), e.getModifiersEx(), e.getKeyCode(), e.getKeyChar(), e.getKeyLocation()));
               return;
            }
         }
      }
   }

   @Override
   public void keyReleased(KeyEvent e) {
   }
}
