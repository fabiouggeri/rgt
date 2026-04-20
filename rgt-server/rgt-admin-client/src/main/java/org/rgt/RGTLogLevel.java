/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import ch.qos.logback.classic.Level;
import ch.qos.logback.classic.Logger;
import ch.qos.logback.classic.LoggerContext;
import ch.qos.logback.classic.encoder.PatternLayoutEncoder;
import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.FileAppender;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public enum RGTLogLevel {
   OFF((byte)0) {
      @Override
      protected Level getLogBackLevel() {
         return Level.OFF;
      }
   },
   ERROR((byte)1) {
      @Override
      protected Level getLogBackLevel() {
         return Level.ERROR;
      }
   },
   WARNING((byte)2) {
      @Override
      protected Level getLogBackLevel() {
         return Level.WARN;
      }
   },
   INFO((byte)3) {
      @Override
      protected Level getLogBackLevel() {
         return Level.INFO;
      }
   },
   DEBUG((byte)4) {
      @Override
      protected Level getLogBackLevel() {
         return Level.DEBUG;
      }
   },
   TRACE((byte)5) {
      @Override
      protected Level getLogBackLevel() {
         return Level.ALL;
      }
   };

   private static RGTLogLevel levelActive = OFF;

   private final byte level;

   private RGTLogLevel(byte level) {
      this.level = level;
   }

   public byte getLevel() {
      return level;
   }

   public boolean isActive() {
      return levelActive.level >= level;
   }

   protected abstract Level getLogBackLevel();

   public void active() {
      final LoggerContext context = (LoggerContext) LoggerFactory.getILoggerFactory();
      if (context != null) {
         final Logger root = context.getLogger(Logger.ROOT_LOGGER_NAME);
         if (root != null) {
            root.setLevel(getLogBackLevel());
            levelActive = this;
         }
      }
   }

   public static RGTLogLevel getByLevel(byte level) {
      for (RGTLogLevel l : values()) {
         if (l.getLevel() == level) {
            return l;
         }
      }
      return OFF;
   }

   public static RGTLogLevel getByName(final String levelName) {
      for (final RGTLogLevel l : values()) {
         if (l.name().equalsIgnoreCase(levelName)) {
            return l;
         }
      }
      return OFF;
   }

   public static RGTLogLevel getLevelActive() {
      return levelActive;
   }

   public static void setLogFileName(final String filePathname) {
      final LoggerContext context = (LoggerContext) LoggerFactory.getILoggerFactory();
      final FileAppender<ILoggingEvent> file = new FileAppender<>();
      final PatternLayoutEncoder layout = new PatternLayoutEncoder();
      final Logger root;

      root = context.getLogger(Logger.ROOT_LOGGER_NAME);
      root.detachAndStopAllAppenders();
      context.reset();

      layout.setContext(context);
      layout.setPattern("%-23d{dd/MM/yyyy HH:mm:ss.SSS} %-5level [%thread:%logger{0}]  %msg%n");
      layout.start();

      file.setName("RGT-SERVER");
      file.setFile(filePathname);
      file.setContext(context);
      file.setAppend(true);
      file.setEncoder(layout);
      file.start();

      root.setLevel(levelActive.getLogBackLevel());
      root.addAppender(file);
      root.setAdditive(false);
   }
}
