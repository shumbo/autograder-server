package admin

// All the API endpoints handled by this package.

import (
    "github.com/eriq-augustine/autograder/api/core"
)

var routes []*core.Route = []*core.Route{
    core.NewAPIRoute(core.NewEndpoint(`admin/course/reload`), HandleCourseReload),
};

func GetRoutes() *[]*core.Route {
    return &routes;
}
