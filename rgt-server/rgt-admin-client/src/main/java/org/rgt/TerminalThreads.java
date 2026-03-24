/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.util.concurrent.ThreadFactory;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalThreads {

   private static final ThreadGroup RGT_THREADS = new ThreadGroup("rgt-threads");

   private static final ThreadGroup RGT_SERVICE_THREADS = new ThreadGroup(RGT_THREADS, "rgt-service-threads");

   private static final ThreadGroup RGT_WORKER_THREADS = new ThreadGroup(RGT_THREADS, "rgt-worker-threads");

   private static final ThreadFactory FACTORY = TerminalThreads::newThread;

   private static final ThreadFactory SERVICE_FACTORY = TerminalThreads::serviceThread;

   private TerminalThreads() {
   }

   public static ThreadFactory factory() {
      return FACTORY;
   }

   public static ThreadFactory serviceFactory() {
      return SERVICE_FACTORY;
   }

   public static Thread newThread(final Runnable r) {
      return new Thread(r);
   }

   public static Thread newThread(final Runnable r, final String name) {
      return new Thread(RGT_WORKER_THREADS, r, name);
   }

   public static Thread newServiceThread(final Runnable r, final String name) {
      return new Thread(RGT_SERVICE_THREADS, r, name);
   }

   public static Thread serviceThread(final Runnable r) {
      return new Thread(RGT_SERVICE_THREADS, r);
   }

   public static Thread[] serviceThreads() {
      final int count = RGT_SERVICE_THREADS.activeCount();
      final Thread threads[] = new Thread[count];
      RGT_SERVICE_THREADS.enumerate(threads);
      return threads;
   }

   public static Thread[] workerThreads() {
      final int count = RGT_WORKER_THREADS.activeCount();
      final Thread threads[] = new Thread[count];
      RGT_THREADS.enumerate(threads);
      return threads;
   }

   public static Thread[] allRgtThreads() {
      final int count = RGT_THREADS.activeCount();
      final Thread threads[] = new Thread[count];
      RGT_THREADS.enumerate(threads);
      return threads;
   }

   public static Thread[] allVMThreads() {
      final int count = Thread.activeCount();
      final Thread threads[] = new Thread[count];
      Thread.enumerate(threads);
      return threads;
   }
}
