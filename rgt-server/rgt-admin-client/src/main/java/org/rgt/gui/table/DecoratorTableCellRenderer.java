/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.awt.Component;
import javax.swing.JLabel;
import javax.swing.JTable;
import javax.swing.table.DefaultTableCellRenderer;

/**
 *
 * @author fabio_uggeri
 */
public class DecoratorTableCellRenderer extends DefaultTableCellRenderer {

   private final TableCellRendererDecorator decorators[];

   public DecoratorTableCellRenderer(TableCellRendererDecorator... decorators) {
      super();
      this.decorators = decorators;
   }

   @Override
   public Component getTableCellRendererComponent(JTable table, Object value, boolean isSelected, boolean hasFocus, int row,
           int column) {
      final JLabel label = (JLabel) super.getTableCellRendererComponent(table, value, isSelected, hasFocus, row, column);
      if (isSelected) {
         label.setBackground(table.getSelectionBackground());
         label.setForeground(table.getSelectionForeground());
      } else {
         label.setBackground(table.getBackground());
         label.setForeground(table.getForeground());
      }
      label.setFont(table.getFont());
      label.setIcon(null);
      for (TableCellRendererDecorator d : decorators) {
         d.decorate(table, table.getModel(), label, row, column, isSelected, hasFocus);
      }
      if (value != null) {
         label.setToolTipText(value.toString());
      }
      return label;
   }

}
