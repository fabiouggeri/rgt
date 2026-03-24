/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.awt.Component;
import java.awt.ComponentOrientation;
import java.awt.Image;
import javax.swing.Icon;
import javax.swing.ImageIcon;
import javax.swing.JLabel;
import javax.swing.JTable;
import javax.swing.SwingConstants;
import javax.swing.table.TableCellRenderer;

/**
 *
 * @author fabio_uggeri
 */
public class CustomTableHeaderRenderer implements TableCellRenderer {

   private final TableCellRenderer renderer;

   private Icon icon = null;

   private final int filterIconPlacement;

   private EditableTableRowSorter rowSorter = null;

   public CustomTableHeaderRenderer(TableCellRenderer renderer, int filterIconPlacement) {
      this.renderer = renderer;
      this.filterIconPlacement = filterIconPlacement;
   }

   public CustomTableHeaderRenderer(TableCellRenderer renderer) {
      this(renderer, SwingConstants.LEADING);
   }

   @Override
   public Component getTableCellRendererComponent(JTable table, Object value, boolean isSelected, boolean hasFocus, int row, int column) {
      final JLabel component = (JLabel) renderer.getTableCellRendererComponent(table, value, isSelected, hasFocus, row, column);
      if (icon != null && isFiltered(column)) {
         Icon newIcon = component.getIcon();
         if (newIcon == null) {
            newIcon = icon;
         } else {
            ComponentOrientation orientation = component.getComponentOrientation();
            if (ComponentOrientation.RIGHT_TO_LEFT.equals(orientation)) {
               if (filterIconPlacement == SwingConstants.LEADING) {
                  newIcon = new CompoundIcon(newIcon, icon);
               } else {
                  newIcon = new CompoundIcon(icon, newIcon);
               }
            } else {
               if (filterIconPlacement == SwingConstants.LEADING) {
                  newIcon = new CompoundIcon(icon, newIcon);
               } else {
                  newIcon = new CompoundIcon(newIcon, icon);
               }
            }
         }
         component.setIcon(newIcon);
      }
      return component;
   }

   public void setIcon(ImageIcon imageIcon) {
      this.icon = new ImageIcon(imageIcon.getImage().getScaledInstance(12, 12, Image.SCALE_SMOOTH));
   }

   public void setIcon(Icon icon) {
      this.icon = icon;
   }

   public void setRowSorter(EditableTableRowSorter rowSorter) {
      this.rowSorter = rowSorter;
   }

   private boolean isFiltered(int column) {
      return rowSorter != null && rowSorter.getRowFilter() != null && rowSorter.getRowFilter().isFiltered(column);
   }
}
