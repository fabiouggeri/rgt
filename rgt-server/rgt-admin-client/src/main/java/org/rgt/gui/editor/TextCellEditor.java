/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.editor;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.util.regex.Pattern;
import javax.swing.Action;
import javax.swing.ActionMap;
import javax.swing.InputMap;
import javax.swing.JButton;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTable;
import javax.swing.JTextArea;
import javax.swing.JTextField;
import javax.swing.KeyStroke;
import javax.swing.SwingUtilities;
import javax.swing.text.AttributeSet;
import javax.swing.text.BadLocationException;
import javax.swing.text.PlainDocument;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 */
public class TextCellEditor extends AbstractCustomCellEditor<JPanel> implements ActionListener {

   private final Pattern pattern;

   private JButton button;

   private JTextField textField;

   private boolean upperCase = false;

   private JTable table = null;

   private int column = -1;

   private boolean showButton = true;

   private String textAreaTitle = null;

   public TextCellEditor() {
      this(false);
   }

   public TextCellEditor(final boolean upperCase) {
      super();
      pattern = null;
      this.upperCase = upperCase;
   }

   public TextCellEditor(final String regulaExpression, final boolean upperCase) {
      super();
      pattern = Pattern.compile(regulaExpression);
   }

   @Override
   public Object getEditorValue() {
      if (textField.getDocument() != null) {
         if (pattern == null || pattern.matcher(textField.getText()).matches()) {
            return textField.getText();
         }
      }
      return null;
   }

   @Override
   public Component getEditorComponent(JTable table, Object value, boolean isSelected, int rowIndex, int colIndex) {
      Component component;
      this.table = table;
      this.column = colIndex;
      component = getComponent();
      if (value == null) {
         textField.setText("");
      } else if (pattern != null && !pattern.matcher(textField.getText()).matches()) {
         textField.setText("");
      } else {
         textField.setText(value.toString());
      }
      textField.setForeground(table.getForeground());
      textField.setBackground(table.getBackground());
      textField.setFont(table.getFont());
      textField.setBorder(TerminalUtil.getTableNoFocusBorder());
      textField.selectAll();
      return component;
   }

   @Override
   public JPanel createComponent() {
      JPanel panel;
      if (showButton) {
         button = new JButton("...");
         button.addActionListener(this);
         button.setFocusable(false);
         button.setFocusPainted(false);
         button.setMargin(new Insets(0, 0, 0, 0));
      }
      textField = new JTextField();
      if (pattern != null) {
         textField.setDocument(new RegExprDocument(pattern, upperCase));
      } else {
         textField.setDocument(new RegExprDocument(upperCase));
      }
      panel = new JPanel(new BorderLayout()) {

         @Override
         public void setBackground(Color bg) {
            textField.setBackground(bg);
         }

         @Override
         public void setForeground(Color fg) {
            textField.setForeground(fg);
         }

         @Override
         public void addNotify() {
            super.addNotify();
            textField.requestFocus();
         }

         @Override
         protected boolean processKeyBinding(KeyStroke ks, KeyEvent e, int condition, boolean pressed) {
            InputMap map = textField.getInputMap(condition);
            ActionMap am = textField.getActionMap();

            if (! textField.isFocusOwner()) {
               textField.requestFocus();
            }
            if (map != null && am != null && isEnabled()) {
               Object binding = map.get(ks);
               Action action = (binding == null) ? null : am.get(binding);
               if (action != null) {
                  if (! textField.isFocusOwner()) {
                     textField.requestFocus();
                  }
                  return SwingUtilities.notifyAction(action, ks, e, textField, e.getModifiersEx());
               }
            }
            return false;
         }
      };
      panel.add(textField);
      if (showButton) {
         panel.add(button, BorderLayout.EAST);
      }
      return panel;
   }

   @Override
   public void actionPerformed(ActionEvent e) {
      int result;
      JTextArea textArea = new JTextArea(10, 50);
      Object value = getEditorValue();
      if (value != null) {
         textArea.setText(((String) value).replace('\u00b6', '\n'));
         textArea.setCaretPosition(0);
      }
      result = JOptionPane.showOptionDialog(null, new JScrollPane(textArea), textAreaTitle(),
              JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, null, null);
      if (result == JOptionPane.OK_OPTION) {
         textField.setText(textArea.getText().replace('\n', '\u00b6'));
      }
   }

   private String textAreaTitle() {
      return textAreaTitle != null ? textAreaTitle : (String) table.getColumnName(column);
   }

   public void setTextAreaTitle(String textAreaTitle) {
      this.textAreaTitle = textAreaTitle;
   }

   /**
    * @return the showButton
    */
   public boolean isShowButton() {
      return showButton;
   }

   /**
    * @param showButton the showButton to set
    */
   public void setShowButton(boolean showButton) {
      this.showButton = showButton;
   }

   @Override
   public void addKeyListener(final KeyListener keyListener) {
      getComponent().addKeyListener(keyListener);
      textField.addKeyListener(keyListener);
   }

   private class RegExprDocument extends PlainDocument {

      private Pattern pattern;

      private boolean upperCase;

      RegExprDocument(final boolean upperCase) {
         this.upperCase = upperCase;
      }

      RegExprDocument(final Pattern pattern, final boolean upperCase) {
         this(upperCase);
         this.pattern = pattern;
      }

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
         if (pattern == null || pattern.matcher(newString).matches()) {
            if (upperCase) {
               super.insertString(offs, str.toUpperCase(), attrSet);
            } else {
               super.insertString(offs, str, attrSet);
            }
         }
      }
   }
}

