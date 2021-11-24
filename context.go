package httprouter

import "context"

type RouteData struct {
	Route  string
	Params map[string]string
}

type routeDataKey struct{}

func WithRouteData(ctx context.Context, data RouteData) context.Context {
	return context.WithValue(ctx, routeDataKey{}, data)
}

func GetRouteData(ctx context.Context) RouteData {
	data, ok := ctx.Value(routeDataKey{}).(RouteData)
	if !ok {
		return RouteData{
			Params: make(map[string]string),
		}
	}
	return data
}

func GetRoute(ctx context.Context) string {
	data := GetRouteData(ctx)

	return data.Route
}

func GetParams(ctx context.Context) map[string]string {
	data := GetRouteData(ctx)

	return data.Params
}
