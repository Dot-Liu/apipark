/*
 * @Date: 2024-02-04 11:09:11
 * @LastEditors: maggieyyy
 * @LastEditTime: 2024-06-04 11:20:24
 * @FilePath: \frontend\packages\core\src\const\team\const.tsx
 */
import { ProColumns } from "@ant-design/pro-components";
import { TeamMemberTableListItem, TeamTableListItem } from "./type";
import { ColumnsType } from "antd/es/table";
import { MemberItem } from "@common/const/type";
import { getItem, getTabItem } from "@common/utils/navigation";
import { SystemTableListItem } from "../system/type";
import { MenuProps, TabsProps } from "antd/lib";
import { Link } from "react-router-dom";

export const TEAM_TABLE_COLUMNS: ProColumns<TeamTableListItem>[] = [
    {
        title: '名称',
        dataIndex: 'name',
        copyable: true,
        ellipsis:true,
        width:160,
        fixed:'left',
        sorter: (a,b)=> {
            return a.name.localeCompare(b.name)
        },
    },
    {
        title: 'ID',
        dataIndex: 'id',
        width: 140,
        copyable: true,
        ellipsis:true
    },
    {
        title: '描述',
        dataIndex: 'description',
        copyable: true,
        ellipsis:true
    },
    {
        title: '服务数量',
        dataIndex: 'systemNum',
        ellipsis:true,
        sorter: (a,b)=> {
            return a.systemNum - b.systemNum
        },
    },
    {
        title: '负责人',
        dataIndex: ['master','name'],
        ellipsis: true,
        width:108,
        filters: true,
        onFilter: true,
        valueType: 'select',
        filterSearch: true,
    },
    {
        title: '创建时间',
        dataIndex: 'createTime',
        ellipsis:true,
        width:176,
        sorter: (a,b)=> {
            return a.createTime.localeCompare(b.createTime)
        },
    },
];


export const TEAM_SYSTEM_TABLE_COLUMNS: ProColumns<SystemTableListItem>[] = [
    {
        title: '服务名称',
        dataIndex: 'name',
        copyable: true,
        ellipsis:true,
        width:160,
        fixed:'left',
        sorter: (a,b)=> {
            return a.name.localeCompare(b.name)
        },
    },
    {
        title: '服务 ID',
        dataIndex: 'id',
        width: 140,
        copyable: true,
        ellipsis:true
    },
    {
        title: '所属组织',
        dataIndex: ['organization','name'],
        copyable: true,
        ellipsis:true
    },
    {
        title: '所属团队',
        dataIndex: ['team','name'],
        copyable: true,
        ellipsis:true
    },
    {
        title: 'API数量',
        dataIndex: 'apiNum',
        ellipsis:true,
        sorter: (a,b)=> {
            return a.apiNum - b.apiNum
        },
    },
    {
        title: '服务数量',
        dataIndex: 'serviceNum',
        ellipsis:true,
        sorter: (a,b)=> {
            return a.serviceNum - b.serviceNum
        },
    },
    {
        title: '负责人',
        dataIndex: ['master','name'],
        ellipsis: true,
        width:108,
        filters: true,
        onFilter: true,
        valueType: 'select',
        filterSearch: true
    },
    {
        title: '添加日期',
        dataIndex: 'createTime',
        ellipsis: true,
        sorter: (a,b)=> {
            return a.createTime.localeCompare(b.createTime)
        },
    },
];

export const TEAM_MEMBER_TABLE_COLUMNS: ProColumns<TeamMemberTableListItem>[] = [
    {
        title: '姓名',
        dataIndex: ['name','name'],
        copyable: true,
        ellipsis:true,
        width:160,
        fixed:'left',
        sorter: (a,b)=> {
            return a.name.name.localeCompare(b.name.name)
        },
    },
    {
        title: '团队角色',
        dataIndex: 'role',
        copyable: true,
        ellipsis:true
    },
    {
        title:'用户组',
        dataIndex:'userGroup',
        ellipsis:true,
        renderText:(_,entity)=>(entity.userGroup?.map((x)=>x.name).join(',')||'-')
    },
    {
        title: '添加日期',
        dataIndex: 'attachTime',
        ellipsis:true,
        sorter: (a,b)=> {
            return a.attachTime.localeCompare(b.attachTime)
        },
    },
];


export const TEAM_MEMBER_MODAL_TABLE_COLUMNS:ColumnsType<MemberItem> = [
    {title:'成员',
    render:(_,entity)=>{
        return <>
            <div>
                <p>
                    <span>{entity.name}</span>
                    {entity.email !== undefined && <span className="text-status_offline">{entity.email}</span>}
                </p>
                <p>{entity.department || '-'}</p>
            </div>
        </>
    }}
]


// export const TEAM_INSIDE_MENU_ITEMS: TabsProps['items'] = [
//     getTabItem(<span>成员管理</span>, 'member',undefined,undefined,'team.myTeam.member.view'),
//     getTabItem(<span>权限管理</span>, 'access',undefined,undefined,'team.myTeam.access.view'),
//     getTabItem(<span>团队设置</span>, 'setting',undefined,undefined,'team.myTeam.self.view')] as TabsProps['items']

    export const TEAM_INSIDE_MENU_ITEMS: MenuProps['items'] = [
        getItem('管理', 'grp', null,
            [
                // getItem(<Link to="system">系统</Link>, 'system'),
                getItem(<Link to="member">成员</Link>, 'member',undefined, undefined, undefined,'team.myTeam.member.view'),
                getItem(<Link to="access">权限</Link>, 'access',undefined,undefined,undefined,'team.myTeam.access.view'),
                getItem(<Link to="setting">设置</Link>, 'setting',undefined,undefined,undefined,['team.myTeam.self.view','system.team.self.edit'])],
            'group'),
    ];
    