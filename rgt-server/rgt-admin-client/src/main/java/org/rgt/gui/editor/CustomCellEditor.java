/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package org.rgt.gui.editor;

import java.awt.event.KeyListener;
import java.util.Comparator;
import javax.swing.table.TableCellEditor;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
public interface CustomCellEditor<T> extends TableCellEditor {

   void addKeyListener(final KeyListener keyListener);

   boolean isChanged();

   boolean isEditable();

   void setEditable(boolean editable);

   void setValueComparator(Comparator comparator);

   public T getComponent();

}

