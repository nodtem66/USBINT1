(function (){
var app = angular.module('WebMonitor', ['ngRoute', 'ngAnimate', 'angularMoment']);

app.controller('AppController', function($rootScope, $scope, $location, $http, $interval, $timeout) {
	
	$scope.config_page = '';
	$scope.isStatusReady = false;
	$rootScope.usbint = [];
	$rootScope.isNotSync = true;
	// stop real-time graph when stream
	$scope.$on('$locationChangeStart', function(event){
		$rootScope.isNotSync = true;
		if ($rootScope.timerRender) $interval.cancel($rootScope.timerRender);
		if ($rootScope.timerQuery) $timeout.cancel($rootScope.timerQuery);
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
  	.success(function(data) {$scope.devices = data.result; $.snackbar({content: "USB already refresh"});});
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
  		$.snackbar({content: "start monitor [" + $scope.pid + "] with device bus-address of [" + bus + ":" + addr + "]"});
  	});
  };
  this.removePatient = function() {}
});

app.controller('HistoryController', ['$scope', function($scope) {

}]);

app.controller('PatientViewController', function($rootScope, $scope, $routeParams, $http, $interval, $timeout){
	var id = $routeParams["patient_id"];
	$scope.patient_id = id
	$scope.params = {};
	$http.get('patient/'+ id +'/mnt').success(function(data){
		$scope.mnts = data.result;
	});
	$scope.dps = [];
	var dps_y = [];
	var buffer_dps = [];
	var dps_max, dps_min;
	var dps_len = 0;

	var chart = new CanvasJS.Chart("myCanvas" , {
		animationEnabled: false,
		backgroundColor: "#000",
		exportEnabled: true,
		interactivityEnabled: false,
		axisX: {
			title: "timeline",
			titleFontSize: 14,
			titleFontColor: "#009eff",
			labelFontSize: 12
		},
		axisY: {
			title: "Voltage",
			titleFontColor: "#009eff",
			titleFontSize: 14,
			labelFontSize: 12
		},
		legend: {
			fontSize: 12,
			fontColor: "#009eff"
		},
		data: [{
			color: "#ff0",
			type: "line",
			xValueType: "dateTime",
			dataPoints: $scope.dps
		}]
	});

	$scope.changeMeasurement = function() {
		var mnt = $scope.select_mnt;
		if (mnt) {
			$http.get('patient/' + id + '/mnt/' + mnt).success(function(data){
				console.log(data);
				$scope.mnt_obj = data.result;
				$scope.params["last_time"] = data.result["last_time"];

				//merge mnt_obj and mnt_tag into mnt_desc
				if ($scope.mnt_tag) {
					$scope.mnt_desc = [];
					var o = data.result.channel_name;
					var t = JSON.parse($scope.mnt_tag[mnt]["description"]);
					for (var i=0,len=o.length; i<len; i++) {
						$scope.mnt_desc[i] = {name: t[i], column: o[i]};
					}
				}
				$.snackbar({content: "get channel list", timeout: 500});
			});
			$http.get('patient/' + id + '/tag').success(function(data){
				var tags = {};
				for (var i=0, len=data.result.length; i < len; i++) {
					var tag = data.result[i];
					tags[tag["mnt"] + '_' + tag["id"]] = tag;
				}
				$scope.mnt_tag = tags;

				// calculate gain for normalize data
				var max = parseInt(tags[mnt]["ref_max"]);
				var min = parseInt(tags[mnt]["ref_min"]);
				var resolution = parseInt(tags[mnt]["resolution"]);
				var sampling_rate = parseInt(tags[mnt]["sampling_rate"]);
				var gain = (max - min) * 1.0 / resolution;
				
				// save normalize parameters
				$scope.params.min = min;
				$scope.params.max = max;
				$scope.params.res = resolution;
				$scope.params.gain = gain;
				$scope.params.rate = sampling_rate;

				//merge mnt_obj and mnt_tag into mnt_desc
				if ($scope.mnt_obj) {
					$scope.mnt_desc = [];
					var o = $scope.mnt_obj.channel_name;
					var t = JSON.parse(tags[mnt].description);
					for(var i=0,len=t.length; i<len; i++) {
						$scope.mnt_desc[i] = {name: t[i],column: o[i]};
					}
				}
			});
		}
	};

	var updateBuffer = function(count, isFirst) {
		var isFirst = isFirst || false;
		var mnt = $scope.select_mnt;
		var tag = $scope["mnt_tag"][mnt];
		var mnt_type = tag.mnt;

		if ($rootScope.isNotSync) return;

		$http.get('patient/' + $scope.patient_id + '/mnt/' + mnt,
			{params: {
				limit: $scope.params.limits, 
				ch: $scope.select_channel,
				after: $scope.params.last_time + 'ns'
			}})
		.success(function(data) {
			var d = data.result;
			var ch = $scope.select_channel;
			var param = $scope.params;

		if (data.result.length == 0) {
			$rootScope.timerQuery = $timeout(updateBuffer, 10);
			return;
		}

		// update buffer
		var last_time = $scope.params.last_time;
		for (var i=d.length-1; i > 0; i--) {
			
			var yy = d[i][ch] * param.gain + param.min;
			if (yy > dps_max) dps_max = yy;
			if (yy < dps_min) dps_min = yy;
			if (last_time < d[i]["time"])
				buffer_dps.push({x: d[i]["time"]/1.0e6, y: yy});
		}
		// update last time
		$scope.params.last_time = d[0]["time"];
		if (mnt_type != 'ecg') {
			dps_min = Math.min.apply(null, dps_y);
			dps_max = Math.max.apply(null, dps_y);
			chart.options.axisY["minimum"] = dps_min;
			chart.options.axisY["maximum"] = dps_max;
		}
		//chart.render();
		//console.log(buffer_dps.length);
		if (isFirst) {
			updateGraph(0, true);
		}
		//if ($rootScope.timerQuery) $timeout.cancel($rootScope.timerQuery);
		$rootScope.timerQuery = $timeout(updateBuffer, 10);
	}).error(function(){$.snackbar({content: 'cannot query data from DB'});});
	};

	var updateGraph = function(counter, isFirst) {
		var isFirst = isFirst || false;
		var limit = $scope.params.limits;
		var mnt = $scope.select_mnt;
		var tag = $scope["mnt_tag"][mnt];
		var mnt_type = tag.mnt;

		// for first time
		if (isFirst) {
			for (var i=0,len=limit; i < len; i++)
			{

				if (buffer_dps.length > 0) {
					var yy = buffer_dps.shift();
					$scope["dps"].push(yy);
					dps_y.push(yy.y);
					if ($scope["dps"].length > limit) {
						$scope["dps"].shift();
						dps_y.shift();
					}
				}
			}
			chart.render();
			if ($rootScope.timerRender) $interval.cancel($scope.timerRender)
			$rootScope.timerRender = $interval(updateGraph, $scope.params.delay);
			return;	
		}
		
		//if (buffer_dps.length > 1000) dps_len += 5;
		//else if (buffer_dps.length > 900) dps_len -= 5;

		//if (buffer_dps.length > 3000) len += 5;
		for (var i=0; i < dps_len; i++)
		{
			if (buffer_dps.length > 0) {
				var yy = buffer_dps.shift();
				$scope["dps"].push(yy);
				dps_y.push(yy.y);
				if ($scope["dps"].length > limit) {
					$scope["dps"].shift();
					dps_y.shift();
				}
			}
		}
		chart.render();
	};
	$scope.refreshGraph = function() {
		// stop update graph
		$rootScope.isNotSync = true;
		if ($rootScope.timerRender) $interval.cancel($rootScope.timerRender);
		if ($rootScope.timerQuery) $timeout.cancel($rootScope.timerQuery);
		
		$scope.dps = [];
		dps_y = [];
		buffer_dps = [];
		chart.options.data[0].dataPoints = $scope.dps;
		chart.render();

		// resize to full height of parent
		var h = $(window).height();
		var elem = $('[auto-height]')
		elem.height(h - elem.offset().top);

		// get tag from mnt
		var mnt = $scope.select_mnt;
		var tag = $scope["mnt_tag"][mnt];
		var mnt_type = tag.mnt;
		
		// config chart option
		if (tag["unit"])
			chart.options["axisY"]["title"] = "Voltage (" + tag["unit"] + ")";
		if (mnt_type == 'ecg') {
			chart.options.axisX.interval = 200;
			chart.options.axisX.intervalType = "millisecond";
			chart.options.axisX.interlacedColor= "#111";
			chart.options.axisX.gridThickness = 1;
			chart.options.axisY.interval = 0.5;
			chart.options.axisY.minimum = -5;
			chart.options.axisY.maximum = 5;
			chart.options.axisY.gridThickness = 1;
			chart.options.axisY.gridColor = "#666";
			$scope.params.rate = $scope.params.rate || 1000;
			$scope.params.limits = 5*$scope.params.rate;
			$scope.params.delay = 100; // milliseconds
			$scope.params.sample_delay = $scope.params.rate * $scope.params.delay / 1000;
			$scope.params.last_time -= 6e9;
		}
		else if (mnt_type == 'oxigen_sat') {
			chart.options.axisY.title = null;			
			chart.options.axisY.labelFontSize = 0;
			chart.options.axisY.labelFontColor = "#000";
			chart.options.axisY.gridThickness = 0;
			$scope.params.rate = $scope.params.rate || 1000;
			$scope.params.limits = 5*$scope.params.rate;
			$scope.params.delay = 25; // milliseconds
			$scope.params.sample_delay = $scope.params.rate / $scope.params.delay;
			$scope.params.last_time -= 5.1e9;
		}
		dps_len = $scope.params.delay;
		$.snackbar({content: "render chart for " + mnt_type});
		$rootScope.isNotSync = false;
		updateBuffer(0, true);
	}
	$scope.changeChannel = function() {
		var channel = $scope.select_channel;
		if (channel) {
			$scope.refreshGraph();
		}
	}
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