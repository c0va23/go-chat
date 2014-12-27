angular
.module("ChatApp.Filters", [])
.filter("parseDate", function(){
  return function(dateStr) {
    return new Date(dateStr);
  };
});

angular
.module("ChatApp.Controllers", [])
.controller("NewMessage", function($scope, $http, $log){
  var user = angular.element("meta[name='user']").attr("content");

  $scope.newMessage ={
    user: user,
    text: "",
  };

  $scope.cleanMessage = function(){
    $scope.newMessage.text = "";
  };

  $scope.send = function() {
    $http
      .post("/api/messages", $scope.newMessage)
      .success(function(){
        $scope.cleanMessage();
      });
  };
})
.controller("Messages", function($scope, $log){
  $scope.messages = [];

  var eventSource = new EventSource("/api/messages");

  eventSource.addEventListener("message", function(messageEvent) {
    var message = angular.fromJson(messageEvent.data);
    $scope.messages.push(message);
    $scope.$apply("messages");

    location.hash = message.id;

    angular.element("#new_message_text").focus();
  });

  eventSource.addEventListener("error", function() {
    $log.error(arguments);
  });
});

angular.module("ChatApp", [
  "ChatApp.Controllers",
  "ChatApp.Filters",
]);
