package constant

var gmsgMap = map[int]string{

	/*
	 * crawl error
	 * 抓取错误
	 */

	//downloader error(下载器错误)
	ERR_CRAWL_DOWNLOADER: "Downloader Error(下载器错误)",
	//analyzer error(分析器错误)
	ERR_CRAWL_ANALYZER: "Analyzer Error(分析器错误)",
	//pipeline error(条目处理器错误)
	ERR_CRAWL_PIPELINE: "Pipeline Error(条目处理器错误)",
	//scheduler error(调度器错误)
	ERR_CRAWL_SCHEDULER: "Scheduler Error(调度器错误)",

	/*
	 * module error
	 */

	//not found module instance(未找到组件实例)
	ERR_MODULE_NOT_FOUND: "Not Found Module Instance(未找到组件实例)",
	//generate MID error(生成MID错误)
	ERR_GENERATE_MID: "Generate MID Error(生成MID错误)",
	//split MID error(拆解mid错误)
	ERR_SPLIT_MID: "Split MID Error(拆解mid错误)",
	//new address error(新建address错误)
	ERR_NEW_ADDRESS: "New Address Error(新建address错误)",
	//register module error(注册module错误)
	ERR_REGISTER_MODULE: "Register Module Error(注册module错误)",
	//illegal module type(非法组件类型)
	ERR_ILLEGAL_MODULE_TYPE: "Illegal Module Type(非法组件类型)",
	//new downloader error(新建下载器失败)
	ERR_NEW_DOWNLOADER_FAIL: "New Downloader Fail(新建下载器失败)",
	//new analyzer error(新建解析器失败)
	ERR_NEW_ANALYZER_FAIL: "New Analyzer Fail(新建解析器失败)",
	//new pipeline error(新建处理管道失败)
	ERR_NEW_PIPELINE: "New Pipeline Fail",
}

func GetErrMsg(errno int) string {
	errmsg, ok := gmsgMap[errno]
	if ok {
		return errmsg
	}
	return ""
}
