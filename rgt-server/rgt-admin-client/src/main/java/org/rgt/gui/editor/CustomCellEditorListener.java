/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package org.rgt.gui.editor;

import javax.swing.event.CellEditorListener;

/**
 *
 * @author fabio_uggeri
 */
public interface CustomCellEditorListener extends CellEditorListener {

   public boolean isValidValue(CustomChangeEvent evt);

   public boolean editStarting(CustomChangeEvent evt);

   public boolean editStopping(CustomChangeEvent evt);

}
