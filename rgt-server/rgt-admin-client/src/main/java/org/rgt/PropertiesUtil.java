/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.io.Reader;
import java.io.Writer;
import java.util.Map.Entry;
import java.util.Properties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class PropertiesUtil {

   private static final Logger LOG = LoggerFactory.getLogger(PropertiesUtil.class);

   public static Properties loadPropertiesFromFile(final String filePathname) {
      return loadPropertiesFromFile(new File(filePathname));
   }
   
   public static Properties loadPropertiesFromFile(final File file) {
      final Properties properties = new Properties();
      Reader fileReader = null;
      try {
         if (file.exists()) {
            fileReader = new FileReader(file);
            properties.load(fileReader);
         } else {
            LOG.warn(file + " not found.");
         }
      } catch (FileNotFoundException ex) {
         LOG.warn(file + " not found.", ex);
      } catch (IOException ex) {
         LOG.error("Error reading preferences from " + file + " not found.", ex);
      } finally {
         if (fileReader != null) {
            try {
               fileReader.close();
            } catch (IOException ex) {
               LOG.error("Error closing file reader for " + file + ".", ex);
            }
         }
      }
      return properties;
   }

   public static boolean savePropertiesToFile(Properties properties, String filePathname, boolean replaceAll) {
      return savePropertiesToFile(properties, new File(filePathname), replaceAll);
   }
   
   public static boolean savePropertiesToFile(Properties properties, File file, boolean replaceAll) {
      Writer fileWriter = null;
      if (! replaceAll) {
         final Properties oldProps = loadPropertiesFromFile(file);
         for (Entry<Object, Object> e : oldProps.entrySet()) {
            if (! properties.containsKey(e.getKey())) {
               properties.put(e.getKey(), e.getValue());
            }
         }
      }
      try {
         fileWriter = new FileWriter(file);
         properties.store(fileWriter, "Remote Graphical Terminal Preferences");
         return true;
      } catch (IOException ex) {
         LOG.error("Error writing configuration file " + file + ".", ex);
      } finally {
         if (fileWriter != null) {
            try {
               fileWriter.close();
            } catch (IOException ex) {
               LOG.error("Error closing file reader for " + file + ".", ex);
            }
         }
      }
      return false;
   }

   public static int getProperty(Properties props, String propName, int defaultValue) {
      final String propValue = props.getProperty(propName);
      if (propValue != null) {
         try {
            return Integer.parseInt(propValue);
         } catch (NumberFormatException ex) {
            LOG.error("{} is an invalid values for '{}': {}.", propValue, propName, ex.getLocalizedMessage());
         }
      }
      return defaultValue;
   }
}
