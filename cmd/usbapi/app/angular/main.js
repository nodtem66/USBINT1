(function (){
var app = angular.module('WebMonitor', ['ngRoute', 'ngAnimate', 'gridster', 'ui.bootstrap']);

// Appcontroller
app.controller('AppController', function($rootScope, $scope, $location, $http, $interval, $timeout) {
	
	$scope.config_page = '';
	$scope.isStatusReady = false;
	$rootScope.usbint = [];
	$rootScope.isNotSync = true;
	$rootScope.timer = [];
	// stop real-time graph when stream
	$scope.$on('$locationChangeStart', function(event){
		$rootScope.isNotSync = true;
		if ($rootScope.timerRender) $interval.cancel($rootScope.timerRender);
		if ($rootScope.timerQuery) $timeout.cancel($rootScope.timerQuery);
		for (var i=0,len=$rootScope.timer.length; i<len; i++) {
			if ($rootScope.timer[i]) {
				if($rootScope.timer[i].render) $interval.cancel($rootScope.timer[i].render);
			}
		}
	});

	$scope.isActive = function(path) {
		var valid_path = $location.path();
		if (valid_path == path) {
			return true;
		}
		return false;
	};
	this.gotoConfigPage = function (page) {
		$scope.config_page = page;
		if (page == 'log') {
			$http.get('sys/print_log').success(function(data) {
				$scope.log_content = data;
			});
		}
		if (page == 'service') {
			$http.get('sys/list_service').success(function(data) {
				$scope.usb_services = data.result;
			});
		}
	};

	$http.get('sys/ip_addr').success(function(data, status) {
		$scope.ipaddr = data.result;
	});
	$http.get('sys/list_process').success(function(data) {
		$scope.process_status = data.result;
		$scope.isStatusReady = true;
	});
	$scope.isPageLoad = true;
});

// TODO: make setting save to file
app.controller('ControlPanelController', function($scope, $http) {
	this.saveSetting = function () {
		$.snackbar({content: "save setting"});
	};
});

app.controller('DashboardController', function($scope, $rootScope, $http) {
	var that = this;
	this.state = 0;
	$http.get("patient")
    .success(function(data) {$scope.names = data.result;});
  
  this.selectUSBPage = function() {
  	$http.get("sys/list_usb")
  	.success(function(data) {$scope.devices = data.result; that.state = 1;});
  };
  this.refreshUSBPage = function() {
  	$http.get("sys/list_usb")
  	.success(function(data) {$scope.devices = data.result; $.snackbar({content: "USB already refresh", timeout: 800});});
  }
  this.selectPatientPage = function(bus_addr) {
  	bus_addr = bus_addr.split(":")
  	this.dev_bus = bus_addr[0] || 0;
  	this.dev_addr = bus_addr[1] || 0;
		this.state = 2;
  };
  this.startNewDevice = function() {
  	var bus = this.dev_bus;
  	var addr = this.dev_addr;
  	$http.get('sys/start', {params: {patient: $scope.pid, bus: bus, addr: addr}})
  	.success(function(data){
  		that.state = 0;
  		$.snackbar({content: "start monitor [" + $scope.pid + "] with device bus-address of [" + bus + ":" + addr + "]", timeout: 800});
  	});
  };
  this.removePatient = function() {}
});

app.controller('HistoryController', ['$scope', function($scope) {

}]);

app.controller('CustomViewController', function($scope){
	var count = 0;
	$scope.gridsterOptions = {
		margins: [20,20], columns: 4, draggable: {handle: 'h3'},
		width: 'auto', 
    colWidth: 'auto',
    rowHeight: 'match'
	};
	$scope.dashboard = {widgets: [{id: count++,col: 0, row: 0, sizeY: 1, sizeX: 1, name: "Widget 1"}]};
	$scope.clear = function() {$scope.dashboard.widgets= [];};
	$scope.addWidget = function(){$scope.dashboard.widgets.push({
		id: count++,
		name: "New Widget",
		sizeX: 1, sizeY: 1
	});};
});


app.controller('WidgetSettingController', function($scope, $http, $modalInstance, widget){
	$http.get('patient').success(function(data){if (data.result) {$scope.patient_ids = data.result}});
	$scope.form = {name: widget.name};
	$scope.submit = function(_form) {
		var f = $scope.form;
		if (!$scope.form.name) {$.snackbar({content: "missing widget name", timeout: 1000}); return;}
		if (!$scope.form.patient) {$.snackbar({content: "please select patient id", timeout: 1000}); return;}
		if (!$scope.form.mnt) {$.snackbar({content: "please select type of measurement", timeout: 1000}); return;}
		if (!$scope.form.ch) {$.snackbar({content: "please select channel name", timeout: 1000}); return;}
		$.snackbar({content: 'load '+f.ch+' for '+f.patient});
		$modalInstance.close($scope.form);

	};
	$scope.changePatient = function() {
		$http.get('patient/'+$scope.form.patient+'/mnt').success(function(data){
			if (data.result) {$scope.mnts = data.result;} else {$.snackbar({content: "cannot load mnt", timeout: 1000});}
		});
	};
	$scope.changeMeasurement = function() {
		$http.get('patient/'+$scope.form.patient+'/mnt/'+$scope.form.mnt)
		.success(function(data){
			if (data.result) {
				$scope.chs = data.result.channel_name;
				var words = $scope.form.mnt.split("_");
				words.splice(words.length - 1, 1);
				var mnt_type = words.join("_");
				if (mnt_type == "oxigen_sat") {
					$scope.chs.push("oxigen_level");
				}	else if (mnt_type == "ecg") {
					$scope.chs.push("heart_rate");
				}
			} else {
				$.snackbar({content: "cannot load channel", timeout: 1000});
			}
		});
	};
});
app.config(['$routeProvider', function($routeProvider, $locationProvider) {
	$routeProvider.
		when('/', {
			templateUrl: 'angular/view/dash_board.html',
			controller: 'DashboardController as d'
		}).
		when('/control_panel', {
			templateUrl: 'angular/view/control_panel.html',
			controller: 'ControlPanelController as c'
		}).
		when('/history', {
			templateUrl: 'angular/view/history.html',
			controller: 'HistoryController as h'
		}).
		when('/custom_view', {
			templateUrl: 'angular/view/custom_view.html',
			controller: 'CustomViewController as c'
		}).
		when('/patient/:patient_id', {
			templateUrl: 'angular/view/patient.html',
			controller: 'PatientViewController as pvc'
		}).otherwise({redirectTo: '/'});
}]);

app.filter('cut', function () {
  return function (value, wordwise, max, tail) {
      if (!value) return '';

      max = parseInt(max, 10);
      if (!max) return value;
      if (value.length <= max) return value;

      value = value.substr(0, max);
      if (wordwise) {
          var lastspace = value.lastIndexOf(' ');
          if (lastspace != -1) {
              value = value.substr(0, lastspace);
          }
      }

      return value + (tail || ' …');
  };
});

}());