/*
 * @Date: 2024-06-04 16:01:08
 * @LastEditors: maggieyyy
 * @LastEditTime: 2024-06-04 19:07:12
 * @FilePath: \frontend\packages\core\src\pages\dashboard\DashboardProjectList.tsx
 */
/*
 * @Date: 2024-06-04 08:54:16
 * @LastEditors: maggieyyy
 * @LastEditTime: 2024-06-04 14:52:44
 * @FilePath: \frontend\packages\core\src\pages\dashboard\DashboardProjectList.tsx
 */
import { useState } from "react";
import { useParams } from "react-router-dom";
import { RouterParams } from "@core/components/aoplatform/RenderRoutes";
import { BasicResponse, STATUS_CODE } from "@common/const/const";
import { SearchBody, MonitorSubscriberData } from "@common/const/dashboard/type";
import { useFetch } from "@common/hooks/http";
import MonitorSubPage, { MonitorSubQueryData } from "@common/components/aoplatform/dashboard/MonitorSubPage";
import DashboardDetail from "./DashboardDetail";
import { TimeRangeButton } from "@common/components/aoplatform/TimeRangeSelector";

export default function DashboardProjectList(){
  const { partitionId,dashboardType } = useParams<RouterParams>()
  const {fetchData } = useFetch()
  const [fullScreen, setFullScreen] = useState<boolean>(false)
  const [queryData, setQueryData] = useState<MonitorSubQueryData>({type:'provider'});
  const [detailId, setDetailId] = useState<string>()
  const [timeButton, setTimeButton] = useState<TimeRangeButton>('hour');
  const [detailEntityName,setDetailEntityName]= useState<string>('')

  const fetchTableData:(body:SearchBody)=>Promise<BasicResponse<{statistics:MonitorSubscriberData[]}>>= (body:MonitorSubQueryData) =>{
    return fetchData<BasicResponse<{statistics:MonitorSubscriberData[]}>>(`monitor/provider`,{
      method:'POST', 
      eoParams:{partition:partitionId},
      eoBody:({...body, dataType:'subscriber'}), 
      eoTransformKeys:['dataType','request_total','request_success','request_rate','proxy_total','proxy_success','proxy_rate','status_fail','avg_resp','max_resp','min_resp','avg_traffic','max_traffic','min_traffic','min_traffic']}).then(resp => {
        if (resp.code === STATUS_CODE.SUCCESS) {
          setQueryData({...body})
          return resp
        }
      })
    }

    
  return (<MonitorSubPage 
              fetchTableData={fetchTableData} 
              timeButton={timeButton} 
              setTimeButton={setTimeButton} 
              detailEntityName={detailEntityName}
              setDetailEntityName={setDetailEntityName}
              detailDrawerContent={<DashboardDetail fullScreen={fullScreen} name={detailEntityName!} queryData={{...queryData,timeButton}} dashboardDetailId={detailId!} partitionId={partitionId!} dashboardType={dashboardType as "api" | "subscriber"}/>} fullScreen={fullScreen} setFullScreen={setFullScreen} setDetailId={setDetailId}/>)
}