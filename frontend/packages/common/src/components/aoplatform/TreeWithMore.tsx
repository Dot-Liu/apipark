/*
 * @Date: 2024-01-31 15:00:10
 * @LastEditors: maggieyyy
 * @LastEditTime: 2024-04-17 16:46:52
 * @FilePath: \frontend\packages\core\src\components\TreeWithMore.tsx
 */
import {CheckOutlined, LoadingOutlined, MoreOutlined} from "@ant-design/icons";
import {Dropdown, Input, InputRef, MenuProps} from "antd";
import { ReactNode, useEffect, useRef, useState} from "react";

export type TreeWithMoreProp = {
    children:ReactNode,
    dropdownMenu:MenuProps['items']
    editable?:boolean
    editingId?:string
    afterEdit?:(val:string)=>Promise<string|boolean>
    editKey?:string
    entity?:{id:string,[k:string]:unknown | string}
    onBlur?:()=>void
}

const TreeWithMore = ({children,dropdownMenu,editable,editingId,entity,editKey='name',afterEdit,onBlur}:TreeWithMoreProp)=>{
    const [editValue, setEditValue] = useState<string>(entity?.[editKey] as string)
    const [submitting, setSubmitting] = useState<boolean>(false)
    const inputRef = useRef<InputRef>(null)

    const handleSubmit = (val:string)=>{
        if(submitting) return
        setSubmitting(true)
        afterEdit && afterEdit(val).finally(()=>setSubmitting(false))
    }

    useEffect(()=>{inputRef.current?.focus()},[inputRef])

    return (<>
        {
        editable  && editingId && entity?.id &&  editingId === entity.id ? <Input ref={inputRef} value={editValue}  onChange={(e)=>{setEditValue(e.target.value)}} onBlur={()=>{onBlur?.()}} onClick={(e)=>e?.stopPropagation()} onPressEnter={()=>{handleSubmit(editValue)}} suffix={submitting ? <LoadingOutlined />:<CheckOutlined onClick={()=>{handleSubmit(editValue)}}/>} />:
        <Dropdown menu={{items:dropdownMenu}}  trigger={['contextMenu']} >
           <div  className='tree-title-hover' >{children}
               <span onClick={(e)=>{ e.stopPropagation();}}>
               <Dropdown menu={{items:dropdownMenu}}  trigger={['click']} >
                    <MoreOutlined  className="tree-title-more" onClick={(e)=>{ e.stopPropagation();}} />
               </Dropdown>
               </span>
           </div>
        </Dropdown>
    }</>)
}
export default TreeWithMore