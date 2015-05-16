(function (){
var app = angular.module('WebMonitor', ['ngRoute', 'ngAnimate', 'angularMoment']);

app.controller('AppController', ['$rootScope', '$location', function($scope, $location) {
	$scope.version = '0.1.0';
	$scope.update_time = '2015-4-12';
	$scope.isPageLoad = false;
	$scope.config_page = '';
	$scope.isActive = function(path) {
		var valid_path = $location.path();
		if (valid_path == path) {
			return true;
		}
		return false;
	};
	this.gotoConfigPage = function (page) {
		$scope.config_page = page;
	};
	//$.material.init();
}]);
app.controller('ControlPanelController', ['$scope', function($scope) {
	$scope.log_content = "test test test";
}]);

app.controller('DashboardController', ['$rootScope','$http', function($rootScope, $http) {
	$http.get("http://localhost:8080/patient")
    .success(function(data) {$rootScope.names = data.result;})
    $http.get("http://localhost:8080/patient/names/tag")
    

}]);

app.controller('HistoryController', ['$scope', function($scope) {

}]);

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
		});
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