/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.awt.Component;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;
import javax.swing.JTable;
import javax.swing.JTextField;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
public abstract class NumberCellEditor<T extends Number> extends AbstractCustomCellEditor<JTextField> {

   protected static final char DECIMAL_SEPARATOR = DecimalFormatSymbols.getInstance().getDecimalSeparator();

   protected static final char GROUP_SEPARATOR = DecimalFormatSymbols.getInstance().getGroupingSeparator();

   protected static final NumberFormat formatter = new DecimalFormat("###,###,###,###,###,###,###,###,##0.##########");

   private final T defaultValue;

   public NumberCellEditor(final T defaultValue) {
      super();
      this.defaultValue = defaultValue;
   }

   @Override
   public Component getEditorComponent(JTable table, Object value, boolean isSelected, int rowIndex, int colIndex) {
      getComponent().setForeground(table.getForeground());
      getComponent().setBackground(table.getBackground());
      getComponent().setFont(table.getFont());
      getComponent().setBorder(TerminalUtil.getTableNoFocusBorder());
      if (value != null) {
         if (value instanceof String) {
            getComponent().setText((String)value);
         } else {
            getComponent().setText(numberToString((T)value));
         }
         getComponent().selectAll();
      }
      return getComponent();
   }

   @Override
   public Object getEditorValue() {
      if (getComponent().getDocument() != null) {
         try {
            return stringtoNumber(getComponent().getText());
         } catch (NumberFormatException e) {
            return null;
         }
      }
      return defaultValue;
   }

   public T getDefaultValue() {
      return defaultValue;
   }

   protected abstract T stringtoNumber(String text);

   protected abstract String numberToString(T value);
}

