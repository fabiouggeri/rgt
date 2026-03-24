/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.table;

import java.awt.Color;
import java.awt.Font;
import javax.swing.Icon;
import javax.swing.JLabel;
import javax.swing.JTable;

/**
 *
 * @author fabio_uggeri
 */
public class ValueDiffDecorator implements TableCellRendererDecorator<ConfigTableModel, JLabel> {

   private Color diffBackground = null;
   private Color diffForeground = null;
   private Icon diffIcon = null;
   private Font diffFont = null;
   private final int compareCol;

   public ValueDiffDecorator(int compareCol) {
      this.compareCol = compareCol;
   }
   
   @Override
   public void decorate(JTable table, ConfigTableModel model, JLabel renderer, int row, int col, boolean isSelected, boolean hasFocus) {
      if (col > 0 && compareCol != col) {
         final int rowModel = table.convertRowIndexToModel(row);
         final int colModel = table.convertColumnIndexToModel(col);
         final int compareColModel = table.convertColumnIndexToModel(compareCol);
         final boolean diff = !model.getValueAt(rowModel, compareColModel).equals(model.getValueAt(rowModel, colModel));
         setRendererBackground(renderer, diff && ! isSelected);
         setRendererForeground(renderer, diff && ! isSelected);
         setRendererFont(renderer, diff);
         setRendererIcon(renderer, diff);
      }
   }

   public Icon diffIcon() {
      return diffIcon;
   }

   public ValueDiffDecorator diffIcon(Icon diffIcon) {
      this.diffIcon = diffIcon;
      return this;
   }

   public Color diffForeground() {
      return diffForeground;
   }

   public ValueDiffDecorator diffForeground(Color diffForeground) {
      this.diffForeground = diffForeground;
      return this;
   }

   public Color diffBackground() {
      return diffBackground;
   }

   public ValueDiffDecorator diffBackground(Color diffBackground) {
      this.diffBackground = diffBackground;
      return this;
   }

   public Font diffFont() {
      return diffFont;
   }

   public ValueDiffDecorator diffFont(Font diffFont) {
      this.diffFont = diffFont;
      return this;
   }

   private void setRendererBackground(JLabel renderer, boolean diff) {
      if (diff && diffBackground != null) {
         renderer.setBackground(diffBackground);
      }
   }

   private void setRendererForeground(JLabel renderer, boolean diff) {
      if (diff && diffForeground != null) {
         renderer.setForeground(diffForeground);
      }
   }

   private void setRendererFont(JLabel renderer, boolean diff) {
      if (diff && diffFont != null) {
         renderer.setFont(diffFont);
      }
   }
   
   private void setRendererIcon(JLabel renderer, boolean diff) {
      if (diff && diffIcon != null) {
         renderer.setIcon(diffIcon);
      }
   }

}
