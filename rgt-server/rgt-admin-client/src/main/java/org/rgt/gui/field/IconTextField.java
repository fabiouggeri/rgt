/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.field;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.RenderingHints;
import javax.swing.Icon;
import javax.swing.JTextField;
import javax.swing.UIManager;

/**
 *
 * @author fabio_uggeri
 */
public class IconTextField extends JTextField {

   private Icon icon;
   private String hint;
   private Insets dummyInsets = null;

   public IconTextField() {
      this(null, "");
   }

   
   public IconTextField(Icon icon, String hint) {
      this.icon = icon;
      this.hint = hint;
   }

   public void setIcon(Icon newIcon) {
      this.icon = newIcon;
   }

   public void setHint(String hint) {
      this.hint = hint;
   }

   @Override
   protected void paintComponent(Graphics g) {
      int textX = 2;
      super.paintComponent(g);

      if (this.dummyInsets == null) {
         this.dummyInsets = UIManager.getBorder("TextField.border").getBorderInsets(this);
      }

      if (this.icon != null) {
         int iconWidth = icon.getIconWidth();
         int iconHeight = icon.getIconHeight();
         int x = dummyInsets.left + 1;
         textX = x + iconWidth + 4;
         int y = (this.getHeight() - iconHeight) / 2;
         icon.paintIcon(this, g, x, y);
      }

      setMargin(new Insets(2, textX, 2, 2));

      if (getText().isEmpty()) {
         int height = this.getHeight();
         Font prev = g.getFont();
         Font italic = prev.deriveFont(Font.ITALIC);
         Color prevColor = g.getColor();
         g.setFont(italic);
         g.setColor(UIManager.getColor("textInactiveText"));
         int h = g.getFontMetrics().getHeight();
         int textBottom = (height - h) / 2 + h - 4;
         int x = this.getInsets().left;
         Graphics2D g2d = (Graphics2D) g;
         RenderingHints hints = g2d.getRenderingHints();
         g2d.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
         g2d.drawString(hint, x, textBottom);
         g2d.setRenderingHints(hints);
         g.setFont(prev);
         g.setColor(prevColor);
      }
   }
}
