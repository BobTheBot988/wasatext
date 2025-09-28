package api

func (rt *_router) PrintNumberOfOpenConnections() {
	rt.baseLogger.Infof("Num of currently open database conn:%d", rt.db.GetNumCurrentlyUsedConn())
	rt.baseLogger.Infof("Num of pool:%d", rt.db.GetNumOpenConn())
}
