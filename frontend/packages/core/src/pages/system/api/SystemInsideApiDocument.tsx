/*
 * @Date: 2024-01-31 15:00:11
 * @LastEditors: maggieyyy
 * @LastEditTime: 2024-06-04 09:55:20
 * @FilePath: \frontend\packages\core\src\pages\system\api\SystemInsideApiDocument.tsx
 */
import  {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import ApiEdit, {ApiEditApi} from "@common/components/postcat/ApiEdit.tsx";
import { Spin, message} from "antd";
import {BasicResponse, STATUS_CODE} from "@common/const/const.ts";
import {useFetch} from "@common/hooks/http.ts";
import { SystemApiDetail, SystemInsideApiDocumentHandle, SystemInsideApiDocumentProps } from "../../../const/system/type.ts";
import { LoadingOutlined } from "@ant-design/icons";


const SystemInsideApiDocument = forwardRef<SystemInsideApiDocumentHandle,SystemInsideApiDocumentProps>((props, ref) => {
    const {systemId, apiId} = props
    const {fetchData} = useFetch()
    const [apiDetail, setApiDetail] = useState<SystemApiDetail>()
    const apiEditRef = useRef<ApiEditApi>(null)
    const [loaded,setLoaded] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)

    useImperativeHandle(ref, ()=>({
        save
    })
)
    useEffect(() => {
        getApiDetail()
    }, []);

    const getApiDetail = ()=>{
        setLoading(true)
        fetchData<BasicResponse<{api:SystemApiDetail}>>('project/api/detail',{method:'GET',eoParams:{project:systemId, api:apiId},eoTransformKeys:['create_time','update_time','match_type','upstream_id','opt_type']}).then(response=>{
            const {code,data,msg} = response
            //console.log(data,code, STATUS_CODE.SUCCESS,code === STATUS_CODE.SUCCESS)
            if(code === STATUS_CODE.SUCCESS){
                setApiDetail(data.api)
                setLoaded(true)
            }else{
                message.error(msg || '操作失败')
            }
        }).finally(()=>{setLoading(false)})
    }

    const save = ()=>{
        return apiEditRef.current?.getData()?.then((res)=>{
            return fetchData<BasicResponse<{id:string}>>('project/api',{method:'PUT',eoParams:{project:systemId,api:apiId},eoBody:(res.apiInfo)}).then(response=>{
                const {code,msg} = response
                if(code === STATUS_CODE.SUCCESS){
                    message.success(msg || '操作成功')
                    return Promise.resolve(true)
                }else{
                    message.error(msg || '操作失败')
                    return Promise.reject(msg|| '操作失败')
                }
            }).catch(errInfo => Promise.reject(errInfo))
        })

    }

    return (<>
        <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} spinning={loading} className=' h-full overflow-auto '>
            <div className="pb-[20px]">
            <ApiEdit apiInfo={apiDetail} editorRef={apiEditRef} loaded={loaded} systemId={systemId}/>
            </div>
        </Spin>
    </>)
})

export default SystemInsideApiDocument