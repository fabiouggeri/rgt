/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;
import java.awt.Rectangle;
import java.awt.Toolkit;
import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.BasicFileAttributes;
import java.nio.file.attribute.FileTime;
import java.text.MessageFormat;
import java.util.Calendar;
import java.util.Collection;
import java.util.Date;
import java.util.Locale;
import java.util.Properties;
import java.util.ResourceBundle;
import javax.swing.UIManager;
import javax.swing.border.Border;
import javax.swing.border.EmptyBorder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalUtil {

   public static final Border NO_FOCUS_CELL_BORDER = new EmptyBorder(0,1,0,1);

   private static final Logger LOG = LoggerFactory.getLogger(TerminalUtil.class);

   private static final int COLUNAS = 2;

   private static final ResourceBundle BUNDLE = ResourceBundle.getBundle("org/rgt/gui/Bundle", Locale.getDefault());

   private static final TerminalUtil INSTANCE = new TerminalUtil();

   private String localVersion = null;

   private TerminalUtil() {
   }

   public static String getMessage(final String name, final Object... args) {
      return MessageFormat.format(BUNDLE.getString(name), args);
   }

   public static boolean hasMessage(final String name) {
      return BUNDLE.containsKey(name);
   }

   public static File getApplicationDir() {
      return new File(System.getProperty("user.dir"));
   }

   public static File getUserHome() {
      return new File(System.getProperty("user.home"));
   }

   public static String absoluteWritablePath(final String pathname) {
      final File filePath = new File(pathname);
      if (!filePath.isAbsolute()) {
         File parent = getApplicationDir();
         if (!Files.isWritable(parent.toPath())) {
            parent = getUserHome();
            LOG.warn("Not allowed to write in {}. Relative path {} changed to writable path {}.",
                    getApplicationDir(), pathname, parent);
         }
         return new File(parent, pathname).getAbsolutePath();
      }
      return pathname;
   }

   public static String getFileExtension(final String fileName) {
      final int index = fileName.lastIndexOf('.');
      /* In Linux hidden directories start with . */
      if (index > 0) {
         return fileName.substring(index + 1);
      } else {
         return "";
      }
   }

   public static String getFileExtension(final File file) {
      return getFileExtension(file.getName());
   }

   public static String getShortestPath(final File path, final File relativeTo) {
      final String originalPath = path.getAbsolutePath();
      final String relativePath = getRelativePath(originalPath, relativeTo.getAbsolutePath());
      return relativePath.length() < originalPath.length() ? relativePath : originalPath;
   }

   public static String getShortestPath(final String pathName, final String relativeTo) {
      final String relativePath = getRelativePath(pathName, relativeTo);
      return relativePath.length() < pathName.length() ? relativePath : pathName;
   }

   public static String getRelativePath(final String pathName, final String relativeTo) {
      final Path p = Path.of(pathName);
      if (p.isAbsolute()) {
         return Path.of(relativeTo).relativize(p).toString();
      } else {
         return pathName;
      }
   }

   public static String getRelativePath(final File pathName, final File relativeTo) {
      final String originalPath = pathName.getAbsolutePath();
      final Path p = Path.of(originalPath);
      if (p.isAbsolute()) {
         return Path.of(relativeTo.getAbsolutePath()).relativize(p).toString();
      } else {
         return originalPath;
      }
   }

   public static String getRelativePath(final String pathName) {
      return getRelativePath(pathName, getApplicationDir().getAbsolutePath());
   }

   public static void centralize(Container window) {
      final GraphicsDevice device = GraphicsEnvironment.getLocalGraphicsEnvironment().getDefaultScreenDevice();
      if (device != null) {
         Rectangle bounds = device.getDefaultConfiguration().getBounds();
         window.setLocation((bounds.width - window.getSize().width) / COLUNAS, (bounds.height - window.getSize().height) / COLUNAS);
      } else {
         final Dimension tela = Toolkit.getDefaultToolkit().getScreenSize();
         window.setLocation((tela.width - window.getSize().width) / COLUNAS, (tela.height - window.getSize().height) / COLUNAS);
      }
   }

   public static void centralize(Container relativeWindow, Container window) {
      if (relativeWindow != null) {
         window.setLocation(relativeWindow.getX() + ((relativeWindow.getWidth() - window.getSize().width) / COLUNAS),
                 relativeWindow.getY() + ((relativeWindow.getHeight() - window.getSize().height) / COLUNAS));
      } else {
         centralize(window);
      }
   }

   public static boolean isEmpty(final String str) {
      return str == null || str.isBlank();
   }

   public static boolean equals(final Object str1, final Object str2) {
      if (str1 == null) {
         return str2 == null;
      } else if (str2 != null) {
         return str1.equals(str2);
      }
      return false;
   }

   public static boolean containsAny(final String str, final String verify) {
      for (int i = 0; i < verify.length(); i++) {
         if (str.indexOf(verify.charAt(i)) >= 0) {
            return true;
         }
      }
      return false;
   }

   public static String getLocalVersion() {
      if (INSTANCE.localVersion == null) {
         INSTANCE.localVersion = INSTANCE.getVersionFromResource("/project.properties");
         if (isEmpty(INSTANCE.localVersion)) {
            INSTANCE.localVersion = INSTANCE.getVersionFromResource("/META-INF/maven/org.rgt/rgt-server/pom.properties");
            if (isEmpty(INSTANCE.localVersion)) {
               INSTANCE.localVersion = "";
            }
         }
      }
      return INSTANCE.localVersion;
   }

   private String getVersionFromResource(final String resourceName) {
      try (final InputStream is = getClass().getResourceAsStream(resourceName)) {
         if (is != null) {
            final Properties properties = new Properties();
            properties.load(is);
            return properties.getProperty("version");
         }
      } catch (IOException ex) {
         LOG.warn("Error getting version from Maven properties", ex);
      }
      return null;
   }

   private static boolean setFileDateTime(final File file, final String attribute, final Date datetime) {
      try {
         final Calendar c = Calendar.getInstance();
         c.setTime(datetime);
         Files.setAttribute(file.toPath(), attribute, FileTime.fromMillis(c.getTimeInMillis()));
         return true;
      } catch (IOException ex) {
         return false;
      }
   }

   public static boolean setCreationTime(final File file, final Date datetime) {
      return setFileDateTime(file, "creationTime", datetime);
   }

   public static boolean setLastAccessTime(final File file, final Date datetime) {
      return setFileDateTime(file, "lastAccessTime", datetime);
   }

   public static boolean setLastModificationTime(final File file, final Date datetime) {
      return setFileDateTime(file, "lastModifiedTime", datetime);
   }

   public static BasicFileAttributes getAttributes(final File file) {
      try {
         return Files.readAttributes(file.toPath(), BasicFileAttributes.class);
      } catch (IOException ex) {
         return null;
      }
   }

   public static String getAbsolutePath(final File file) {
      try {
         return file.getCanonicalPath();
      } catch (IOException ex) {
         return file.getAbsolutePath();
      }
   }

   public static boolean sleep(final long time) {
      try {
         Thread.sleep(time);
         return true;
      } catch (InterruptedException ex1) {
         return false;
      }
   }

   public static String toString(final Collection<File> files, final String sep) {
      final StringBuilder sb = new StringBuilder(files.size() * 15);
      files.forEach(f -> {
         if (sb.length() > 0) {
            sb.append(sep);
         }
         sb.append(f.getName());
      });
      return sb.toString();
   }

   public static Border getTableNoFocusBorder() {
      Border border = UIManager.getBorder("Table.cellNoFocusBorder");
      if (border != null) {
         return border;
      }
      return NO_FOCUS_CELL_BORDER;
   }
   
  public static String getTextClipboard() {
    Transferable t = Toolkit.getDefaultToolkit().getSystemClipboard().getContents(null);
    try {
      if (t != null && t.isDataFlavorSupported(DataFlavor.stringFlavor)) {
        String text = (String) t.getTransferData(DataFlavor.stringFlavor);

        return text.trim();
      }
    } catch (UnsupportedFlavorException | IOException e) {
       return "";
    }
    return "";
  }   
}
