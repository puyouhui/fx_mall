<template>
  <view class="policy-page">
    <!-- 自定义导航栏 -->
    <view class="custom-header">
      <view class="navbar-fixed" style="background-color: #20CB6B;">
        <!-- 状态栏撑起高度 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>
        <!-- 导航栏内容区域 -->
        <view class="navbar-content" :style="{ height: navBarHeight + 'px' }">
          <view class="navbar-left" @click="goBack">
            <uni-icons type="left" size="20" color="#fff"></uni-icons>
          </view>
          <view class="navbar-title">
            <text class="navbar-title-text">{{ pageTitle }}</text>
          </view>
          <view class="navbar-right"></view>
        </view>
      </view>
    </view>

    <!-- 占位符，避免内容被导航栏遮挡 -->
    <view :style="{ height: (statusBarHeight + navBarHeight) * 2 + 'rpx' }"></view>

    <!-- 正文 -->
    <scroll-view scroll-y class="content-scroll">
      <view class="content-container">
        <!-- 隐私政策 -->
        <view v-if="policyType === 'privacy'">
          <view class="section-title">一、总则</view>
          <text class="paragraph">
            橙心选（云南）供应链管理有限公司（以下简称“本公司”或“我们”）非常重视您的个人信息和隐私保护。
            您在使用橙心选小程序服务（包括但不限于商品浏览、下单、支付、售后等功能）时，我们将按照本《隐私政策》收集、
            使用和保护您的相关信息。本政策旨在向您清晰说明我们如何处理您的信息，以及您享有的权利。
          </text>

          <view class="section-title">二、我们如何收集与使用您的信息</view>
          <text class="paragraph">
            1. 账号信息：为为您创建账户或识别您的身份，我们可能会收集您的手机号、昵称、联系方式等基础信息，用于登录验证、
            安全风控、订单通知等场景。
          </text>
          <text class="paragraph">
            2. 订单与配送信息：当您在小程序中下单时，我们会收集与订单相关的信息，包括收货地址、联系人、联系电话、商品信息、
            订单金额、支付状态等，用于完成订单处理、配送服务、售后服务及必要的账务对账。
          </text>
          <text class="paragraph">
            3. 设备与日志信息：为保障系统安全、改进产品体验，我们会依法合规记录设备型号、操作系统、网络环境、访问时间、
            页面停留情况等日志信息，仅用于统计分析、性能优化与故障排查，不会用于识别您的个人身份。
          </text>
          <text class="paragraph">
            4. 客服与反馈信息：当您通过客服或意见反馈与我们联系时，我们可能会保存您主动提供的描述、图片、联系方式等信息，
            以便及时跟进和问题追踪。
          </text>

          <view class="section-title">三、Cookie 和同类技术</view>
          <text class="paragraph">
            为使您获得更便捷的访问体验，我们可能会在小程序内使用 Cookie 或类似技术缓存少量必要信息，
            用于记录登录状态、偏好设置或协助统计分析。您可以通过相关设置进行清理，但可能会影响部分功能的使用体验。
          </text>

          <view class="section-title">四、我们如何共享、转让和公开披露信息</view>
          <text class="paragraph">
            1. 我们不会向任何无关第三方出售您的个人信息，仅在以下情形下可能共享： \n  
            （1）为实现配送、支付结算等业务功能，必要时与物流服务商、支付机构等合作伙伴共享订单相关信息； \n  
            （2）基于法律法规、行政或司法机关要求，依法对外提供相关信息； \n  
            （3）为维护您或其他用户、公众生命财产等重大合法权益，在紧急情况下合理共享。
          </text>
          <text class="paragraph">
            2. 除非取得您的明确授权或依法依规要求，我们不会向任何第三方转让您的个人信息。
          </text>
          <text class="paragraph">
            3. 如需进行匿名化统计分析，我们会对数据进行去标识化处理，确保无法指向特定自然人。
          </text>

          <view class="section-title">五、我们如何保护您的信息</view>
          <text class="paragraph">
            我们采用合理可行的安全防护措施，包括但不限于访问控制、数据加密、审计日志等技术手段，
            以防止信息被未经授权访问、使用或泄露。同时，我们会对员工、合作伙伴进行安全与合规培训，
            要求其在授权范围内使用数据，并承担相应的保密义务。
          </text>

          <view class="section-title">六、您的权利</view>
          <text class="paragraph">
            在法律法规允许的范围内，您有权对个人信息进行查询、更正、补充或删除，也可以申请注销账号或撤回授权。
            部分权利可以通过小程序内的“系统设置”“个人资料”等入口自助完成，您也可以通过页面提供的联系方式与我们取得联系。
          </text>

          <view class="section-title">七、未成年人保护</view>
          <text class="paragraph">
            本服务主要面向具有完全民事行为能力的用户。如您为未成年人，请在监护人同意并在其指导下使用本服务。
            如监护人发现我们在未获同意的情况下收集了未成年人的信息，可与我们联系协助删除。
          </text>

          <view class="section-title">八、本政策的更新</view>
          <text class="paragraph">
            随着业务发展或法律法规变化，我们可能适时对本《隐私政策》进行更新。
            重要变更时，我们将以合理方式（包括但不限于站内公告、弹窗提示等）进行提醒，并在更新生效后继续为您提供服务。
          </text>

          <view class="section-title">九、联系我们</view>
          <text class="paragraph">
            如您对本政策有任何疑问、意见或建议，可通过“关于我们”页面中公示的联系方式与我们取得联系，
            我们会尽快为您处理。
          </text>
        </view>

        <!-- 用户协议 -->
        <view v-else>
          <view class="section-title">一、协议声明</view>
          <text class="paragraph">
            本《用户协议》（以下简称“本协议”）是您（以下简称“用户”或“您”）与橙心选（云南）供应链管理有限公司
            就使用橙心选小程序相关服务所达成的法律协议。您访问或使用本小程序，即视为您已阅读、理解并同意受本协议约束。
          </text>

          <view class="section-title">二、账户注册与使用</view>
          <text class="paragraph">
            1. 您应当保证提供的注册信息真实、准确、完整，并在信息变更时及时更新。
            因您提供信息不真实或未及时更新而导致的后果，由您自行承担。
          </text>
          <text class="paragraph">
            2. 账号仅限您本人或您授权的人员在合法、合规的场景下使用。您应妥善保管账号和密码，
            不得出租、出借、出售或以其他方式转让给第三方使用。因保管不善造成的风险和损失，由您自行承担。
          </text>

          <view class="section-title">三、商品与订单</view>
          <text class="paragraph">
            1. 小程序展示的商品信息（包括但不限于名称、规格、图片、价格、库存等）仅为交易邀请，
            不构成本公司的要约。您的下单行为构成订单申请，订单经系统确认并生成后，方视为交易成立。
          </text>
          <text class="paragraph">
            2. 由于市场变化、系统故障或其他因素，可能导致个别商品价格、库存等信息异常。
            如订单信息明显异常或无法履行，我们有权在与您沟通后进行调整、取消订单或协商其他解决方案。
          </text>
          <text class="paragraph">
            3. 您应在下单前仔细核对收货地址、联系方式、配送时间等信息。因您填写错误导致的延误、
            退货或额外费用，由您自行承担。
          </text>

          <view class="section-title">四、配送与收货</view>
          <text class="paragraph">
            1. 我们将通过自营或合作物流的方式为您提供配送服务。配送范围、时效和费用以页面展示或双方另行约定为准。
          </text>
          <text class="paragraph">
            2. 收货时，请您或指定收货人对商品数量、外包装完整性等进行验收。
            如发现异常，请在签收时当场记录并及时与我们联系，逾期可能影响您的售后权益。
          </text>

          <view class="section-title">五、售后与退款</view>
          <text class="paragraph">
            1. 我们会依据国家相关法律法规及平台规则，为符合条件的订单提供退换货或退款服务。
          </text>
          <text class="paragraph">
            2. 鉴于纸品、餐饮一次性耗材等商品的特殊性，一经拆封或影响二次销售的商品，除质量问题外一般不支持无理由退货。
          </text>

          <view class="section-title">六、用户行为规范</view>
          <text class="paragraph">
            您在使用本服务过程中，不得从事任何违法违规或有违公序良俗的行为，包括但不限于： \n  
            （1）发布或传播违法违规内容； \n  
            （2）对本平台进行逆向工程、恶意攻击或干扰正常运营； \n  
            （3）利用本平台从事虚假交易、洗钱等违法行为； \n  
            （4）侵犯任何第三方的合法权益（包括但不限于知识产权、名誉权等）。
          </text>

          <view class="section-title">七、服务变更、中断与终止</view>
          <text class="paragraph">
            我们有权根据业务调整或运营需要，对服务内容进行变更、暂停或终止。
            如涉及重大变更，我们会通过合理方式事先通知您。因不可抗力、网络故障、政策变化等因素导致的服务中断，
            我们将尽力修复，但在法律允许范围内对由此产生的间接损失不承担责任。
          </text>

          <view class="section-title">八、责任限制</view>
          <text class="paragraph">
            在法律允许的范围内，本公司对因以下原因造成的损失不承担责任： \n  
            （1）由于您自身原因导致的账号、密码泄露或信息发布； \n  
            （2）第三方服务（如支付机构、物流服务商等）自身故障； \n  
            （3）不可抗力或意外事件。
          </text>

          <view class="section-title">九、适用法律与争议解决</view>
          <text class="paragraph">
            本协议适用中华人民共和国法律。因本协议或使用本服务所产生的争议，双方应友好协商解决；
            协商不成的，可向本公司所在地有管辖权的人民法院提起诉讼。
          </text>

          <view class="section-title">十、其他</view>
          <text class="paragraph">
            本协议条款如因部分无效或被撤销，并不影响其他条款的效力。未尽事宜，您可通过“关于我们”页面公示的联系方式与我们沟通确认。
          </text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      statusBarHeight: 20,
      navBarHeight: 45,
      policyType: 'privacy' // 'privacy' | 'terms'
    };
  },
  computed: {
    pageTitle() {
      return this.policyType === 'privacy' ? '隐私政策' : '用户协议';
    }
  },
  onLoad(options) {
    const systemInfo = uni.getSystemInfoSync();
    this.statusBarHeight = systemInfo.statusBarHeight || 20;

    // 处理传入参数：type=privacy | terms
    const type = (options && options.type) || 'privacy';
    this.policyType = type === 'terms' ? 'terms' : 'privacy';

    // 自定义导航高度
    try {
      const menuButtonInfo = uni.getMenuButtonBoundingClientRect();
      if (menuButtonInfo) {
        this.navBarHeight =
          (menuButtonInfo.bottom - this.statusBarHeight) +
          (menuButtonInfo.top - this.statusBarHeight);
      }
    } catch (e) {
      // 忽略，使用默认高度
    }
  },
  methods: {
    goBack() {
      uni.navigateBack();
    }
  }
};
</script>

<style scoped>
.policy-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.custom-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.navbar-fixed {
  background-color: #20CB6B;
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30rpx;
}

.navbar-left,
.navbar-right {
  width: 60rpx;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.navbar-title {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
}

.content-scroll {
  height: calc(100vh - 0px);
}

.content-container {
  padding: 24rpx 30rpx 60rpx 30rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-top: 32rpx;
  margin-bottom: 16rpx;
}

.paragraph {
  display: block;
  font-size: 26rpx;
  color: #666;
  line-height: 1.8;
  text-align: justify;
  margin-bottom: 16rpx;
}
</style>

