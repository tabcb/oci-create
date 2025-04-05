func Oracle_Account_Manage(content *gin.Context) {

	if !Request_Auth(content.Query("user"), content.Query("key"), content.ClientIP()) {
		return
	}

	// 初始化
	Oracle_Section = Oracle_Section_Row[Current_Profile_Num-1]
	Oracle_Section_Name = Oracle_Section.Name()
	log.Printf("[Info] [%s] 云账户管理正在初始化...", Oracle_Section_Name)
	ctx = context.Background()
	_, _, _, _, IdentityClient, _, Oracle_Account_Info, status := Oracle_Account_Init_Var(Oracle_Section)

	// 如果初始化失败
	if status {
		log.Printf("[Info] [%s] 云账户管理初始化失败.", Oracle_Section_Name)

		// Api 返回
		var Users []string
		data := map[string]interface{}{
			"msg":   "failed",
			"code":  0,
			"sum":   "",
			"users": Users,
		}
		content.JSON(200, data)

		// 消息推送
		line1 := "*#云账户管理*\n"
		line2 := " *[ " + Oracle_Section_Name + " ] 云账户管理初始化失败:*\n"
		line3 := "(a) 账号无管理账户权限;\n"
		line4 := "(b) 账户已封号."
		text := line1 + line2 + line3 + line4

		// 上报数据
		go utils.Notice_Send(text)

		// 结束
		return
	}

	// 列取用户 || 重置密码 || 详细信息
	if content.Query("action") == "list" || content.Query("action") == "reset" || content.Query("action") == "detail" {

		// 创建一个用户列取请求
		Oracle_User_List_Request := identity.ListUsersRequest{
			CompartmentId: common.String(Oracle_Account_Info.Tenancy),
		}

		// 调用身份管理 API 列取用户
		Oracle_User_List_Result, err := IdentityClient.ListUsers(ctx, Oracle_User_List_Request)
		if err != nil {

			log.Printf("[Info] [%s] 用户列取失败:%s", Oracle_Section_Name, err.Error())
			var Users []string
			data := map[string]interface{}{
				"msg":   "failed",
				"code":  0,
				"sum":   "",
				"users": Users,
			}
			content.JSON(200, data)
			// 消息推送
			line1 := "*#云账户管理*\n"
			line2 := " *[ " + Oracle_Section_Name + " ] 用户列取失败*"
			text := line1 + line2

			// 上报数据
			go utils.Notice_Send(text)

			// 结束
			return
		} else {
			log.Printf("[Info] [%s] 用户列取成功.", Oracle_Section_Name)
		}

		// 合并用户到一个组
		var Users []string
		for _, user := range Oracle_User_List_Result.Items {
			Users = append(Users, *user.Name)
		}

		// 列取用户
		if content.Query("action") == "list" {
			data := map[string]interface{}{
				"msg":   "success",
				"code":  0,
				"sum":   len(Users),
				"users": Users,
			}
			content.JSON(200, data)

			// 结束
			return
		}
		if content.Query("action") == "reset" {
			num, err := strconv.Atoi(content.Query("id"))
			if err != nil {
				return
			}
			Oracle_User_Password_Reset_Request := identity.CreateOrResetUIPasswordRequest{
				UserId: common.String(*Oracle_User_List_Result.Items[num-1].Id),
			}

			Oracle_User_Password_Reset_Response, err2 := IdentityClient.CreateOrResetUIPassword(context.Background(), Oracle_User_Password_Reset_Request)
			if err2 != nil {
				log.Printf("[Info] [%s] 为用户 %s 重置密码失败：%s", Oracle_Section_Name, *Oracle_User_List_Result.Items[num-1].Name, err2.Error())

				// 消息推送
				line1 := "*#云账户管理*\n"
				line2 := " *[ " + Oracle_Section_Name + " ] 账户重置密码失败:*\n"
				line3 := "(a) 账户为 IDCS. IDCS 账户不支持直接重置密码；\n"
				line4 := "(b) 账户为 SSO. SSO 联盟用户不支持重置密码，可选择重置 OCI 账户密码."
				text := line1 + line2 + line3 + line4

				// 上报数据
				go utils.Notice_Send(text)

				// 结束
				return
			}
			log.Printf("[Info] [%s] 为用户 %s 重置密码成功", Oracle_Section_Name, *Oracle_User_List_Result.Items[num-1].Name)

			// 消息推送
			line1 := "*#云账户管理*\n"
			line2 := " *[ " + Oracle_Section_Name + " ] 用户重置密码成功*\n"
			line3 := "用户: `" + *Oracle_User_List_Result.Items[num-1].Name + "` \n"
			line4 := "密码: `" + *Oracle_User_Password_Reset_Response.Password + "`"
			text := line1 + line2 + line3 + line4

			// 上报数据
			go utils.Notice_Send(text)

			// Api 返回
			data := map[string]interface{}{
				"msg":  "success",
				"code": 0,
			}
			content.JSON(200, data)

			// 结束
			return
		}
		if content.Query("action") == "detail" {
			num, err := strconv.Atoi(content.Query("id"))
			if err != nil {
				return
			}
			text := "*#云账户管理*\n "
			text = text + " *[ " + Oracle_Section_Name + " ] *\n"
			// 由于出现一些问题，用for获取数据
			for i := 1; i < 10; i++ {
				if i == 1 {
					if Oracle_User_List_Result.Items[num-1].Name != nil {
						line := "*用户               ：*`" + strings.Replace(fmt.Sprintln(*Oracle_User_List_Result.Items[num-1].Name), "\n", "", -1) + "`\n"
						text = text + line
					}
				} else if i == 2 {
					if Oracle_User_List_Result.Items[num-1].Email != nil {
						line := "*邮箱               ：*`" + strings.Replace(fmt.Sprintln(*Oracle_User_List_Result.Items[num-1].Email), "\n", "", -1) + "`\n"
						text = text + line
					}
				} else if i == 3 {
					if Oracle_User_List_Result.Items[num-1].Description != nil {
						line := "*描述               ：*`" + strings.Replace(fmt.Sprintln(*Oracle_User_List_Result.Items[num-1].Description), "\n", "", -1) + "`\n"
						text = text + line
					}
				} else if i == 4 {
					line := "*状态               ：*`" + strings.Replace(fmt.Sprintln(Oracle_User_List_Result.Items[num-1].LifecycleState), "\n", "", -1) + "`\n"
					text = text + line
				} else if i == 5 {
					if Oracle_User_List_Result.Items[num-1].LastSuccessfulLoginTime != nil {
						line := "*上次登录时间：*`" + Oracle_User_List_Result.Items[num-1].LastSuccessfulLoginTime.String() + "`\n"
						text = text + line
					} else {
						line := "*上次登录时间：*`暂无`" + "\n"
						text = text + line
					}
				} else if i == 6 {
					if Oracle_User_List_Result.Items[num-1].TimeCreated != nil {
						line := "*用户创建时间：*`" + Oracle_User_List_Result.Items[num-1].TimeCreated.String() + "`\n"
						text = text + line
					} else {
						line := "*用户创建时间：*`暂无`" + "\n"
						text = text + line
					}
				}
			}

			// 上报数据
			text = text + ". "
			go utils.Notice_Send(text)
		}
	}
	// 新用户
	if content.Query("action") == "new" {

		// 创建一个新的用户请求
		Oracle_User_Create_Request := identity.CreateUserRequest{
			CreateUserDetails: identity.CreateUserDetails{
				Name:        common.String(content.Query("email")),
				Email:       common.String(content.Query("email")),
				Description: common.String(content.Query("email")),
			},
		}
		Oracle_User_Create_Request.CompartmentId = common.String(Oracle_Account_Info.Tenancy)

		// 调用身份管理 API 创建用户
		Oracle_User_Create_Result, err := IdentityClient.CreateUser(ctx, Oracle_User_Create_Request)
		if err != nil {
			log.Printf("[Info] [%s] 用户 %s 创建失败:%s", Oracle_Section_Name, content.Query("email"), err.Error())

			// 消息推送
			line1 := "*#云账户管理*\n"
			line2 := " *[ " + Oracle_Section_Name + " ] 添加用户失败：*"
			line3 := err.Error()
			text := line1 + line2 + line3

			// 上报数据
			go utils.Notice_Send(text)

			return
		}
		log.Printf("[Info] [%s] 用户 %s 创建成功.", Oracle_Section_Name, content.Query("email"))

		// 列取 Group
		Oracle_Group_List_Request := identity.ListGroupsRequest{
			CompartmentId: common.String(Oracle_Account_Info.Tenancy),
		}

		// 发送请求
		Oracle_Group_List_Result, err := IdentityClient.ListGroups(context.Background(), Oracle_Group_List_Request)
		if err != nil {
			log.Printf("[Info] [%s] 列取组失败:%s", Oracle_Section_Name, err.Error())

			// 消息推送
			line1 := "*#云账户管理*\n"
			line2 := " *[ " + Oracle_Section_Name + " ] 添加用户失败：*"
			line3 := err.Error()
			text := line1 + line2 + line3

			// 上报数据
			go utils.Notice_Send(text)

			return
		}

		// 添加组
		for _, group := range Oracle_Group_List_Result.Items {
			log.Printf("[Info] [%s] 用户 %s 正在添加入组 %s ...", Oracle_Section_Name, content.Query("email"), *group.Name)
			time.Sleep(5 * time.Second)
			request := identity.AddUserToGroupRequest{
				AddUserToGroupDetails: identity.AddUserToGroupDetails{
					GroupId: common.String(*group.Id),
					UserId:  common.String(*Oracle_User_Create_Result.User.Id),
				},
			}

			// 发送请求
			_, err := IdentityClient.AddUserToGroup(context.Background(), request)
			if err != nil {
				log.Printf("[Info] [%s] 为用户 %s 添加入组 %s 失败：%s", Oracle_Section_Name, content.Query("email"), *group.Name, err.Error())
			}

		}
		log.Printf("[Info] [%s] 用户 %s 创建成功且已经添加所有组.", Oracle_Section_Name, content.Query("email"))

		// 重置密码
		Oracle_User_Password_Reset_Request := identity.CreateOrResetUIPasswordRequest{
			UserId: common.String(*Oracle_User_Create_Result.User.Id),
		}
		Oracle_User_Password_Reset_Response, err2 := IdentityClient.CreateOrResetUIPassword(context.Background(), Oracle_User_Password_Reset_Request)
		if err2 != nil {
			log.Printf("[Info] [%s] 为用户 %s 创建密码失败：%s", Oracle_Section_Name, content.Query("email"), err2.Error())

			// 消息推送
			line1 := "*#云账户管理*\n"
			line2 := " *[ " + Oracle_Section_Name + " ] 添加用户成功*\n"
			line3 := "用户: `" + content.Query("email") + "` \n"
			line4 := "*已经向您的邮箱发送了一封密码重置邮件，请注意查收.*"

			text := line1 + line2 + line3 + line4

			// 上报数据
			go utils.Notice_Send(text)

			return
		}
		log.Printf("[Info] [%s] 为用户 %s 创建密码成功", Oracle_Section_Name, content.Query("email"))

		// 消息推送
		line1 := "*#云账户管理*\n"
		line2 := " *[ " + Oracle_Section_Name + " ] 添加用户成功*\n"
		line3 := "用户: `" + content.Query("email") + "` \n"
		line4 := "密码: `" + *Oracle_User_Password_Reset_Response.Password + "` "
		text := line1 + line2 + line3 + line4

		// 上报数据
		go utils.Notice_Send(text)
	}

}