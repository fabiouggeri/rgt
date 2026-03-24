/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import java.util.List;
import org.rgt.TerminalServer;

/**
 *
 * @author fabio_uggeri
 */
public interface RemoteFileOperationListener {

   void listFiles(final TerminalServer server, final String folderPathname, final List<FileInfo> file);
   void removeFile(final TerminalServer server, final FileInfo file);
   void uploadFile(final TerminalServer server, FileInfo source, FileInfo target);
   void downloadFile(final TerminalServer server, FileInfo source, FileInfo target);
   void notification(final String msg);
}
