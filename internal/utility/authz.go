package utility

import (
	"errors"

	"golershop.cn/internal/consts"
	"golershop.cn/internal/model"
)

// RequireLogin 校验登录用户存在。
func RequireLogin(user *model.ContextUser) error {
	if user == nil {
		return errors.New("请先登录")
	}
	return nil
}

// RequireRole 校验用户角色。
func RequireRole(user *model.ContextUser, roleId uint, deniedMsg string) error {
	if err := RequireLogin(user); err != nil {
		return err
	}
	if user.RoleId != roleId {
		if deniedMsg == "" {
			deniedMsg = "无权限操作"
		}
		return errors.New(deniedMsg)
	}
	return nil
}

// CheckStoreScope 校验 store_id 数据权限。
// 规则：平台/租户放行；商家仅可操作自己的店铺。
func CheckStoreScope(user *model.ContextUser, targetStoreId uint) error {
	if err := RequireLogin(user); err != nil {
		return err
	}
	if user.RoleId == consts.ROLE_ADMIN || user.RoleId == consts.ROLE_SITE {
		return nil
	}
	if user.RoleId == consts.ROLE_SELLER && user.StoreId == targetStoreId {
		return nil
	}
	return errors.New("无权限操作")
}

// CheckChainScope 校验 chain_id 数据权限。
// 规则：平台/租户放行；门店仅可操作自己的门店。
func CheckChainScope(user *model.ContextUser, targetChainId uint) error {
	if err := RequireLogin(user); err != nil {
		return err
	}

	if user.RoleId == consts.ROLE_ADMIN || user.RoleId == consts.ROLE_SITE {
		return nil
	}

	if user.RoleId == consts.ROLE_CHAIN && user.ChainId == targetChainId {
		return nil
	}
	return errors.New("拒绝访问！")
}
