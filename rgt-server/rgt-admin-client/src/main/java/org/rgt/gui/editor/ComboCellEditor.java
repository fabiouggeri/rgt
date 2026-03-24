/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.awt.Component;
import java.awt.event.FocusAdapter;
import java.awt.event.FocusEvent;
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import javax.swing.JComboBox;
import javax.swing.JTable;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
public class ComboCellEditor<T> extends AbstractCustomCellEditor<JComboBox> {

   private final T[] values;

   public ComboCellEditor(T[] values) {
      super();
      this.values = values;
   }

   @Override
   public Component getEditorComponent(JTable table, Object value, boolean isSelected, int rowIndex, int vColIndex) {
      if (value == null) {
         getComponent().setSelectedIndex(0);
      } else {
         getComponent().setSelectedItem(value);
         getComponent().requestFocus();
      }
      return getComponent();
   }

   @Override
   public Object getEditorValue() {
      return getComponent().getSelectedItem();
   }

   @Override
   public JComboBox createComponent() {
      JComboBox component = new JComboBox(values);
      component.setSelectedIndex(0);
      component.addKeyListener(new KeyAdapter() {
         @Override
         public void keyPressed(KeyEvent e) {
            if (e.getComponent() instanceof JComboBox) {
               JComboBox comp = (JComboBox) e.getComponent();
               if (!comp.isPopupVisible()) {
                  comp.setPopupVisible(true);
               }
            }
         }
      });
      component.addFocusListener(new FocusAdapter() {
         @Override
         public void focusGained(FocusEvent e) {
            if (e.getComponent() instanceof JComboBox) {
               JComboBox comp = (JComboBox) e.getComponent();
               if (!comp.isPopupVisible()) {
                  comp.setPopupVisible(true);
               }
            }
         }
      });
      return component;
   }
}
