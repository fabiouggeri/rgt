/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.util.HashMap;
import java.util.Map;
import javax.swing.ImageIcon;
import javax.swing.JLabel;
import javax.swing.JTable;
import javax.swing.SwingConstants;
import org.rgt.TerminalUtil;
import org.rgt.protocol.admin.files.FileInfo;

/**
 *
 * @author fabio_uggeri
 */
public class FileExtensionDecorator implements TableCellRendererDecorator<FileInfoModel, JLabel> {

   private final ImageIcon anyIcon = new ImageIcon(getClass().getResource("/file_any_16.png"));

   private final ImageIcon folderIcon = new ImageIcon(getClass().getResource("/folder_16.png"));

   private final Map<String, ImageIcon> extensionsMap = new HashMap<>();

   public FileExtensionDecorator() {
      extensionsMap.put("log", new ImageIcon(getClass().getResource("/file_log_16.png")));
      extensionsMap.put("txt", new ImageIcon(getClass().getResource("/file_txt_16.png")));
      extensionsMap.put("exe", new ImageIcon(getClass().getResource("/file_exe_16.png")));
   }

   @Override
   public void decorate(JTable table, FileInfoModel model, JLabel label, int row, int col, boolean isSelected, boolean hasFocus) {
      final FileInfo file = model.fileInfoAt(table.convertRowIndexToModel(row));
      if (file == null) {
         return;
      }
      label.setHorizontalTextPosition(SwingConstants.TRAILING);
      if (file.getFileType() == FileInfo.FileType.DIRECTORY) {
         label.setIcon(folderIcon);
      } else {
         final String extension = TerminalUtil.getFileExtension(file.getName()).toLowerCase();
         final ImageIcon icon = extensionsMap.get(extension);
         if (icon != null) {
            label.setIcon(icon);
         } else {
            label.setIcon(anyIcon);
         }
      }
   }

}
