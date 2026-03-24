/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.text.ParseException;
import javax.swing.JTextField;

/**
 *
 * @author fabio_uggeri
 */
public class IntegerCellEditor extends NumberCellEditor<Long> {

   public IntegerCellEditor(final long defaultValue) {
      super(defaultValue);
   }

   @Override
   protected Long stringtoNumber(String text) {
      try {
         return formatter.parse(text).longValue();
      } catch (ParseException ex) {
         return 0l;
      }
   }

   @Override
   protected String numberToString(Long value) {
      return formatter.format(value.longValue()).replace(Character.toString(GROUP_SEPARATOR), "");
   }

   @Override
   public JTextField createComponent() {
      IntTextField component = new IntTextField();
      component.setText(numberToString(getDefaultValue()));
      return component;
   }
}
