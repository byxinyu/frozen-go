@extends('layout.main')

@section('title','系统设置')

@section('content')

    <br /><br />
    <h2 class="ui dividing header">
        <i class="settings icon"></i>
        <div class="content">
            首次设置
            <div class="sub header">接下来将设置一些重要内容，这个过程不会很长</div>
        </div>
    </h2>
    <br />
    <div class="ui negative message">
        <i class="close icon"></i>
        <div class="header">
            重要提示！
        </div>
        <p>本页面可多次打开重复设置，请记录本页面域名：{{ url('/firstuse') }}
        </p>
    </div>
    <h3 class="ui top attached header">
        管理员设置
    </h3>
    <form class="ui settings form" method="post" action="{{ url('/firstregis') }}">
        {{ csrf_field() }}
    <div class="ui attached segment">
        @if(!$issetuser)
            <div class="field">
                <label>用户名</label>
                <input type="text" name="username" placeholder="设置管理员用户名">
            </div>
            <div class="field">
                <label>密码</label>
                <input type="password" name="password" placeholder="设置管理员密码">
            </div>
    </div>
        @else
            <p>管理员信息请登录面板后台后修改！</p>
        @endif
    <h3 class="ui attached header">
        Daemon信息设置
    </h3>
    <div class="ui attached segment">
        <div class="field">
            <label>Daemon连接IP</label>
            <input type="number" name="ip" placeholder="请输入安装了Daemon的服务器ip，本机请填127.0.0.1">
        </div>
        <div class="field">
            <label>Daemon连接端口</label>
            <input type="number" name="port" placeholder="请输入Daemon服务器的端口，默认为52023">
        </div>
    </div>
    <h3 class="ui attached header">
        其他设置
    </h3>
    <div class="ui attached segment">
        <p>以下设置仅限专业人员改动，如您不清楚以下设置项有何用处请留空！</p>
    </div>
    <h3 class="ui bottom attached header">
        <button class="ui primary button" type="submit">保存设置</button>
    </h3>
    </form>
    <br />

    <script type="text/javascript">
        $(function(){
            $('.settings.form').form({
                inline:'true',
                fields:{
                    username:{
                        identifier:'username',
                        rules:[{
                            type:'empty',
                            prompt:'用户名不得为空！'
                        }]
                    },
                    password:{
                        identifire:'password',
                        rules:[{
                            type:'empty',
                            prompt:'密码不得为空！'
                        }]
                    },
                    ip:{
                        identifire:'ip',
                        rules:[{
                            type:'empty',
                            prompt:'Daemon连接ip不得为空！'
                        }]
                    },
                    port:{
                        identifire:'port',
                        rules:[{
                            type:'empty',
                            prompt:'Daemon连接端口不得为空！'
                        }]
                    },
                }
            });
        });
    </script>
@endsection