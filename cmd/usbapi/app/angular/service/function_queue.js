(function () {
  //Start CLOSURE
  var app = angular.module('mainApp');
  /* ---------------- MessageQueue ------------------------------------------------------ */
  app.service('function_queue', function messageQueueFactory($injector, ngProgress) {

    var working_queue = [];
    var waiting_queue = [];
    var total_task = 0;
    var done_task = 0;
    var self = this;
    var isComplete = false;
    var isEnable = false;
    var isWaiting = false;
    var currentProgress = 0;
    var isDenied = false;

    this.enable = function(set) {
      isEnable = set;
    };

    this.clear = function() {
      if (!isEnable || isWaiting)
        return;
      total_task = 0;
      done_task = 0;
    };

    this.isEmpty = function() {
      return working_queue.length === 0 && waiting_queue.length === 0;
    };

    this.acceptResult = function() {
      waiting_queue.shift();
      done_task++;
      isWaiting = false;
    };

    this.denyResult = function() {
      var task = waiting_queue.shift();
      working_queue.push(task);
      isWaiting = false;
      isDenied = true;
    };

    this.push = function(lambda) {
      if (typeof lambda === "function") {
        working_queue.push(lambda);
        total_task++;
        isComplete = false;
        //console.log("add");
      }
    };

    this.process = function() {
      if (!isEnable)
        return;
      if (working_queue.length > 0 && waiting_queue.length === 0) {
        var task = working_queue.shift();
        waiting_queue.push(task);
        if (typeof task !== "function") {
          return;
        }
        //console.log("process [" + done_task + "/" + total_task + "]");
        $injector.invoke(task, self);
        if (!isDenied) {
          currentProgress = 0;
        }
        isWaiting = true;
        isDenied = false;
      }
      //Update Progress
      if (total_task == 0) 
        return;
      var percent = 100.0 * done_task / total_task;
      var k = Math.pow(total_task, -0.5);
      if (percent > currentProgress) {
        currentProgress = percent;
      }
      if (done_task < total_task) {
        if (currentProgress + k < 100 * (done_task + 1) / total_task) {
          currentProgress += k;
        }
      }
      if (percent >= 100) {
        if (!isComplete) {
          ngProgress.complete();
          isComplete = true;
          isWaiting = false;
        }
      } else {
        ngProgress.set(currentProgress);
        isComplete = false;
      }
      //console.log(percent, currentProgress, done_task, total_task);
    };
  });
  //END CLOSURE
})();