angular.module('WebMonitor').controller('WidgetController', function($rootScope, $scope, $modal, $http, $timeout, $interval){

	$scope.params = {};
	$scope.dps = [];
	$rootScope.timer[$scope.widget.id] = {};
	var notChart = {"oxigen_level": 1, "heart_rate": 1};
	var dps_y = [0,0,0,0];
	var dps_x1 = [0,0,0,0];
	var dps_x2 = [0,0,0,0];
	var buffer_dps = [0,0,0,0];
	var dps_max, dps_min;
	var dps_len = 0;
	var chart;
	
	var createChart = function(id){
		chart = new CanvasJS.Chart("widget_"+id , {
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
	});};
	var resizeChart = function() {
		if (chart != null) {
			$timeout(function(){
			var elem = $("#widget_"+$scope.widget.id);
			chart.options.width = elem.width() - 24;
			chart.options.height = elem.height() - 48;
			chart.render();
			}, 1000);
		}
	};

	$scope.$watch('widget.sizeX', resizeChart, true);
	$scope.$watch('widget.sizeY', resizeChart, true);
	$scope.remove = function(widget){
		$scope.dashboard.widgets.splice($scope.dashboard.widgets.indexOf(widget), 1);
	};

	$scope.openSettings = function(widget) {
		$modal.open({
			animation: $scope.animationsEnabled,
			scope: $scope,
			templateUrl: 'angular/view/widget_setting.html',
			controller: 'WidgetSettingController as w',
			resolve: {widget: function() {return widget;}}
		}).result.then(function(params){
			widget.name = params.name;
			$scope.conf = params;
			if (!notChart[params.ch]) {
				createChart(widget.id);
				resizeChart(0, 1);
			} else {
				 $('#widget_' + $scope.widget.id).css({"font-size": "6em", "color": "#008AFF", "text-align": "center"});
			}
			// query mnt data
			$http.get('patient/' + params.patient + '/mnt/' + params.mnt).success(function(data){
				$scope.mnt_obj = data.result;
				$scope.params["last_time"] = data.result["last_time"];

				//merge mnt_obj and mnt_tag into mnt_desc
				if ($scope.mnt_tag) {
					$scope.mnt_desc = [];
					var o = data.result.channel_name;
					var t = JSON.parse($scope.mnt_tag[params.mnt]["description"]);
					for (var i=0,len=o.length; i<len; i++) {
						$scope.mnt_desc[i] = {name: t[i], column: o[i]};
					}
					//refresh chart
					if (!notChart[params.ch]) refreshGraph();
					else {$rootScope.isNotSync = false;updateBuffer(0, true);}
				}
			});
			// query tag data
			$http.get('patient/' + params.patient + '/tag').success(function(data){
				var tags = {};
				var mnt = params.mnt;
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
				if (notChart[params.ch]) {
					$scope.params.rate = $scope.params.rate || 1000;
					$scope.params.limits = 5*$scope.params.rate;
					$scope.params.delay = 25; // milliseconds
					$scope.params.sample_delay = $scope.params.rate / $scope.params.delay;
					$scope.params.last_time -= 5.1e9;
				}
				
				//merge mnt_obj and mnt_tag into mnt_desc
				if ($scope.mnt_obj) {
					$scope.mnt_desc = [];
					var o = $scope.mnt_obj.channel_name;
					var t = JSON.parse(tags[mnt].description);
					for(var i=0,len=t.length; i<len; i++) {
						$scope.mnt_desc[i] = {name: t[i],column: o[i]};
					}
					//refresh chart
					if (!notChart[params.ch]) refreshGraph();
					else {
						$rootScope.isNotSync = false;
						updateBuffer(0, true);
					}
				}
			});
		});
	};

	var updateBuffer = function(count, isFirst) {
		var isFirst = isFirst || false;
		var mnt = $scope.conf.mnt;
		var tag = $scope["mnt_tag"][mnt];
		var mnt_type = tag.mnt;

		if ($rootScope.isNotSync) return;

		var channel = $scope.conf.ch;

		if (channel == 'oxigen_level') {
			channel = "led1,led2";
			$http.get('patient/' + $scope.conf.patient + '/mnt/' + $scope.conf.mnt,
				{params: {
					limit: $scope.params.limits, 
					ch: channel,
					after: $scope.params.last_time + 'ns'
				}})
			.success(function(data) {
				var d = data.result;
				var ch = channel;
				var param = $scope.params;

			if (data.result.length == 0) {
				$rootScope.timer[$scope.widget.id].query = $timeout(updateBuffer, 100);
				return;
			}

			// update buffer
			var last_time = $scope.params.last_time;
			for (var i=d.length-1; i > 0; i--) {
				
				var yy1 = d[i]["led1"] * param.gain + param.min;
				var yy2 = d[i]["led2"] * param.gain + param.min;

				var leny1 = dps_y.length;
				var leny2 = buffer_dps.length;
				// store raw data
				dps_x1.push(yy1);
				if (dps_x1.length > 4) dps_x1.shift();
				dps_x2.push(yy2);
				if (dps_x2.length > 4) dps_x2.shift();

				// filter lowpass 10Hz order 3 Fs=1000
				yy1 = dps_x1[3] + 3*dps_x1[2] + 3*dps_x1[1] + dps_x1[0] + 2.8744*dps_y[leny1-1] - 2.7565*dps_y[leny1-2] + 0.8819*dps_y[leny1-3];
				yy2 = dps_x2[3] + 3*dps_x2[2] + 3*dps_x2[1] + dps_x2[0] + 2.8744*buffer_dps[leny2-1] - 2.7565*buffer_dps[leny2-2] + 0.8819*buffer_dps[leny2-3];
				dps_y.push(yy1);
				buffer_dps.push(yy2)

				// oxigen level calculation
				// calculate AC/DC
				var ac1, dc1, ac2, dc2;
				if (dps_y.length > 2000) {
					dps_y.shift();
				}
				
				if (buffer_dps.length > 2000) {
					buffer_dps.shift();
				}
			}			
			dps_min = Math.min.apply(null, buffer_dps);
			dps_max = Math.max.apply(null, buffer_dps);
			ac2 = dps_max - dps_min;
			dc2 = (dps_max + dps_min)/2;
		
			dps_min = Math.min.apply(null, dps_y);
			dps_max = Math.max.apply(null, dps_y);
			ac1 = dps_max - dps_min;
			dc1 = (dps_max + dps_min)/2;

			var r = (ac2/dc2)/(ac1/dc1);
			r = Math.round(-22.5667*r*r - 0.8653*r + 105.7353);
			//console.log(ac1, ac2);
			if (ac1 < 500 || ac2 < 500) {
				$('#widget_' + $scope.widget.id).text("--");	
			} else if (r < 0 || r > 100) {
				$('#widget_' + $scope.widget.id).text("--");	
			} else {
				$('#widget_' + $scope.widget.id).text(r + '%');
			}
			// update last time
			$scope.params.last_time = d[0]["time"];			
			$rootScope.timer[$scope.widget.id] = $timeout(updateBuffer, 100);

	}).error(function(){$.snackbar({content: 'cannot query data from DB'});});

		} else if (channel == 'heart_rate') {
			channel = "lead_ii";
		} else {
		$http.get('patient/' + $scope.conf.patient + '/mnt/' + $scope.conf.mnt,
			{params: {
				limit: $scope.params.limits, 
				ch: channel,
				after: $scope.params.last_time + 'ns'
			}})
		.success(function(data) {
			var d = data.result;
			var ch = channel;
			var param = $scope.params;

		if (data.result.length == 0) {
			$rootScope.timer[$scope.widget.id].query = $timeout(updateBuffer, 100);
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
		if (isFirst) {
			updateGraph(0, true);
		}
		
		$rootScope.timer[$scope.widget.id] = $timeout(updateBuffer, 100);

	}).error(function(){$.snackbar({content: 'cannot query data from DB'});});}
	};

	var updateGraph = function(counter, isFirst) {
		var isFirst = isFirst || false;
		var limit = $scope.params.limits;
		var mnt = $scope.conf.mnt;
		var tag = $scope["mnt_tag"][mnt];
		var mnt_type = tag.mnt;

		// for first time
		if (isFirst) {
			for (var i=0,len=limit; i < len; i++)
			{
				if (buffer_dps.length > 0) {
					var yy = buffer_dps.shift();
					yy.y *= -1;
					$scope["dps"].push(yy);
					dps_y.push(yy.y);
					if ($scope["dps"].length > limit) {
						$scope["dps"].shift();
						dps_y.shift();
					}
				}
			}
			chart.render();
			if ($rootScope.timer[$scope.widget.id].render) $interval.cancel($rootScope.timer[$scope.widget.id].render)
			$rootScope.timer[$scope.widget.id].render = $interval(updateGraph, $scope.params.delay);
			return;	
		}
		
		for (var i=0; i < dps_len; i++)
		{
			if (buffer_dps.length > 0) {
				var yy = buffer_dps.shift();
				yy.y *= -1;
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

	var refreshGraph = function() {
		// stop update graph
		$rootScope.isNotSync = true;
		if ($rootScope.timer[$scope.widget.id].render) $interval.cancel($rootScope.timer[$scope.widget.id].render);
		
		$scope.dps = [];
		dps_y = [];
		buffer_dps = [];
		chart.options.data[0].dataPoints = $scope.dps;
		
		// resize to full height of parent
		var elem = $("#widget_"+$scope.widget.id);
		chart.options.width = elem.width() - 24;
		chart.options.height = elem.height() - 48;
		chart.render();

		// get tag from mnt
		var mnt = $scope.conf.mnt;
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
		$rootScope.isNotSync = false;
		updateBuffer(0, true);
	}
});
