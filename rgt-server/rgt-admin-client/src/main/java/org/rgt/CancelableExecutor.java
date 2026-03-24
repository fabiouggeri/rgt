/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package org.rgt;

import java.util.concurrent.Executor;

/**
 *
 * @author fabio_uggeri
 */
public interface CancelableExecutor extends Executor, Runnable {

   public void stop(boolean force);
   
   public void stop();
   
   public void clear();
   
}
