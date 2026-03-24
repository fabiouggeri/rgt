/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.table;

import java.awt.Color;
import java.awt.Font;
import java.awt.font.TextAttribute;
import java.util.Collections;
import javax.swing.Icon;
import javax.swing.JLabel;
import javax.swing.JTable;

/**
 *
 * @author fabio_uggeri
 */
public class ModifiedCellDecorator implements TableCellRendererDecorator<ConfigTableModel, JLabel> {

   private Color changedBackground = null;
   private Color changedForeground = null;
   private Icon changedIcon = null;
   private Font changedFont = null;

   @Override
   public void decorate(JTable table, ConfigTableModel model, JLabel renderer, int row, int col, boolean isSelected, boolean hasFocus) {
      if (col == 0) {
         if (model.isChanged(row)) {
            renderer.setFont(renderer.getFont().deriveFont(Collections.singletonMap(TextAttribute.WEIGHT, TextAttribute.WEIGHT_BOLD)));
         }
      } else {
         final int rowModel = table.convertRowIndexToModel(row);
         final int colModel = table.convertColumnIndexToModel(col);
         final boolean changed = model.isChanged(rowModel, colModel);
         setRendererBackground(renderer, changed && ! isSelected);
         setRendererForeground(renderer, changed && ! isSelected);
         setRendererFont(renderer, changed);
         setRendererIcon(renderer, changed);
      }
   }

   public Icon changedIcon() {
      return changedIcon;
   }

   public ModifiedCellDecorator changedIcon(Icon changedIcon) {
      this.changedIcon = changedIcon;
      return this;
   }

   public Color changedForeground() {
      return changedForeground;
   }

   public ModifiedCellDecorator changedForeground(Color changedForeground) {
      this.changedForeground = changedForeground;
      return this;
   }

   public Color changedBackground() {
      return changedBackground;
   }

   public ModifiedCellDecorator changedBackground(Color changedBackground) {
      this.changedBackground = changedBackground;
      return this;
   }

   public Font changedFont() {
      return changedFont;
   }

   public ModifiedCellDecorator changedFont(Font changedFont) {
      this.changedFont = changedFont;
      return this;
   }

   private void setRendererBackground(JLabel renderer, boolean changed) {
      if (changed && changedBackground != null) {
         renderer.setBackground(changedBackground);
      }
   }

   private void setRendererForeground(JLabel renderer, boolean changed) {
      if (changed && changedForeground != null) {
         renderer.setForeground(changedForeground);
      }
   }

   private void setRendererFont(JLabel renderer, boolean changed) {
      if (changed && changedFont != null) {
         renderer.setFont(changedFont);
      }
   }
   
   private void setRendererIcon(JLabel renderer, boolean changed) {
      if (changed && changedIcon != null) {
         renderer.setIcon(changedIcon);
      }
   }
}
