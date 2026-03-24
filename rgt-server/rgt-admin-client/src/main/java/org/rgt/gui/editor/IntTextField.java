/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.util.regex.Pattern;
import javax.swing.JTextField;
import javax.swing.text.AttributeSet;
import javax.swing.text.BadLocationException;
import javax.swing.text.Document;
import javax.swing.text.PlainDocument;

/**
 *
 * @author fabio_uggeri
 */
public class IntTextField extends JTextField {

   private static final Pattern SIGNED_PATTERN = Pattern.compile("\\-[0-9]*");

   private static final Pattern UNSIGNED_PATTERN = Pattern.compile("[0-9]+");

   private boolean unsigned = false;

   public IntTextField() {
      super();
   }

   @Override
   protected Document createDefaultModel() {
      return new IntTextDocument();
   }

   public long longValue() {
      if (getDocument() != null) {
         try {
            return Long.parseLong(getText());
         } catch (NumberFormatException e) {
         }
      }
      return 0L;
   }

   public int intValue() {
      if (getDocument() != null) {
         try {
            return Integer.parseInt(getText());
         } catch (NumberFormatException e) {
         }
      }
      return 0;
   }

   public short shortValue() {
      if (getDocument() != null) {
         try {
            return Short.parseShort(getText());
         } catch (NumberFormatException e) {
         }
      }
      return 0;
   }

   public byte byteValue() {
      if (getDocument() != null) {
         try {
            return Byte.parseByte(getText());
         } catch (NumberFormatException e) {
         }
      }
      return 0;
   }

   public boolean isUnsigned() {
      return unsigned;
   }

   public void setUnsigned(boolean unsigned) {
      this.unsigned = unsigned;
   }

   private class IntTextDocument extends PlainDocument {

      @Override
      public void insertString(int offs, String str, AttributeSet attrSet) throws BadLocationException {
         String newString;
         if (str == null) {
            return;
         }
         if (getLength() == 0) {
            newString = str;
         } else {
            final String oldString = getText(0, getLength());
            newString = oldString.substring(0, offs) + str + oldString.substring(offs);
         }
         if (unsigned) {
            if (UNSIGNED_PATTERN.matcher(newString).matches()) {
               super.insertString(offs, str, attrSet);
            }
         } else {
            if (UNSIGNED_PATTERN.matcher(newString).matches() || SIGNED_PATTERN.matcher(newString).matches()) {
               super.insertString(offs, str, attrSet);
            }
         }
      }
   }
}
