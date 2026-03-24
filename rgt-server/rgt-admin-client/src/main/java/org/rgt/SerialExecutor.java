/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.util.concurrent.ArrayBlockingQueue;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class SerialExecutor implements CancelableExecutor {

   private static final Logger LOG = LoggerFactory.getLogger(SerialExecutor.class);

   private static final Runnable END_TASK = () -> {
   };
   
   private Thread thread = null;
   
   private final String name;

   private final ArrayBlockingQueue<Runnable> tasks = new ArrayBlockingQueue<>(512);

   public static synchronized SerialExecutor newInstance(final String name) {
      final SerialExecutor exe = new SerialExecutor(name);
      exe.start();
      return exe;
   }

   private SerialExecutor(final String name) {
      this.name = name;
   }

   @Override
   public void run() {
      try {
         Runnable runnable = nextTask();
         while (runnable != null && runnable != END_TASK) {
            runnable.run();
            runnable = nextTask();
         }
      } finally {
         thread = null;
      }

   }

   private Runnable nextTask() {
      try {
         return tasks.take();
      } catch (InterruptedException | RuntimeException ex) {
         LOG.error("Error getting next task", ex);
      }
      return null;
   }

   @Override
   public void execute(Runnable command) {
      tasks.offer(command);
   }

   public synchronized void start() {
      if (thread == null) {
         thread = TerminalThreads.newThread(this, name != null ? name : "RGT Serial Executor");
      }
      thread.start();
   }

   @Override
   public void stop(boolean force) {
      if (force) {
         clear();
      }
      stop();
   }

   @Override
   public void stop() {
      tasks.offer(END_TASK);
   }

   @Override
   public void clear() {
      tasks.clear();
   }
}
